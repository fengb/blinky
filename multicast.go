package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"time"
)

const sendInterval = 30 * time.Second

type Multicast struct {
	conf          *Conf
	snapshotState *SnapshotState
	aes           *Aes
	listen        *AutoListen
	sendTimer     *time.Timer
}

func NewMulticast(conf *Conf, snapshotState *SnapshotState) (*Multicast, error) {
	m := Multicast{
		conf:          &Conf{},
		snapshotState: snapshotState,
		sendTimer:     time.NewTimer(1 * time.Nanosecond),
	}

	err := m.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range m.sendTimer.C {
			err := m.send()
			if err != nil {
				log.Println(err)
			}
			m.sendTimer.Reset(sendInterval)
		}
	}()

	return &m, nil
}

func (m *Multicast) UpdateConf(conf *Conf) error {
	var (
		aes    *Aes
		listen *AutoListen
	)

	err := Parallel(
		func() (err error) {
			aes, err = NewAes(conf.Multicast.Secret)
			return err
		},
		func() (err error) {
			listen, err = m.updateListen(conf)
			return err
		},
	)

	if err != nil {
		closeAll(listen)
		return err
	}

	if m.listen != listen {
		closeAll(m.listen)
		m.listen = listen
	}

	if !conf.Multicast.Send {
		m.sendTimer.Stop()
	}

	m.aes = aes
	m.conf = conf

	return nil
}

func closeAll(closers ...io.Closer) {
	for _, closer := range closers {
		if closer == nil || reflect.ValueOf(closer).IsNil() {
			continue
		}

		err := closer.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

func (m *Multicast) updateListen(conf *Conf) (*AutoListen, error) {
	if !conf.Multicast.Listen {
		return nil, nil
	}

	if conf.Multicast.Listen == m.conf.Multicast.Listen &&
		conf.Multicast.Addr == m.conf.Multicast.Addr {
		return m.listen, nil
	}

	addr, err := net.ResolveUDPAddr("udp", conf.Multicast.Addr)
	if err != nil {
		return nil, err
	}

	listener, err := NewAutoListen(addr)
	if err != nil {
		return nil, err
	}

	go func() {
		for msg := range listener.C {
			err := m.recv(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return listener, nil
}

func (m *Multicast) recv(msg ReadMsg) error {
	//if m.send != nil && m.send.LocalAddr().String() == msg.Src.String() {
	//	// We sent this. Ignore!
	//	continue
	//}

	snapshot, err := m.decode(msg.Packet)
	if err != nil {
		return err
	}

	ipString := msg.Src.IP.String()
	host, err := net.LookupAddr(ipString)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s — %s", ipString, host)
	m.snapshotState.UpdateRemote(key, snapshot)
	return nil
}

func (m *Multicast) send() error {
	addr, err := net.ResolveUDPAddr("udp", m.conf.Multicast.Addr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	snapshotData, err := m.encode(m.snapshotState.Local())
	if err != nil {
		return err
	}

	_, err = conn.Write(snapshotData)
	if err != nil {
		return err
	}

	return nil
}

func (m *Multicast) decode(encrypted []byte) (*Snapshot, error) {
	compressed, err := m.aes.Decrypt(encrypted)
	if err != nil {
		// Wrong secret. Might be a bug?
		return nil, err
	}

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		// Bad compression. Not a bug?
		return nil, err
	}
	defer gz.Close()
	plaintext, err := ioutil.ReadAll(gz)
	if err != nil {
		// Bad compression. Not a bug?
		return nil, err
	}

	snapshot := Snapshot{}
	err = json.Unmarshal(plaintext, &snapshot)
	if err != nil {
		// Wrong data structure. Probably a bug
		return nil, err
	}

	return &snapshot, nil
}

func (m *Multicast) encode(snapshot *Snapshot) ([]byte, error) {
	plaintext, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	compressed := bytes.Buffer{}
	gz := gzip.NewWriter(&compressed)
	_, err = gz.Write(plaintext)
	gz.Close()
	if err != nil {
		return nil, err
	}

	ciphertext, err := m.aes.Encrypt(compressed.Bytes())
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}
