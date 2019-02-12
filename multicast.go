package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"
)

type Multicast struct {
	conf          *Conf
	snapshotState *SnapshotState
	aes           *Aes
	listen        *AutoListen
	sendTimer     *time.Timer
}

func NewMulticast(conf *Conf, snapshotState *SnapshotState) (m *Multicast, err error) {
	m = &Multicast{
		conf:          conf,
		snapshotState: snapshotState,
		sendTimer:     time.NewTimer(1 * time.Nanosecond),
	}

	m.aes, err = NewAes(conf.Multicast.Secret)
	if err != nil {
		return nil, err
	}

	m.listen, err = m.initListen(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range m.sendTimer.C {
			err := m.send()
			if err != nil {
				log.Println(err)
			}
			m.sendTimer.Reset(m.conf.Multicast.Send)
		}
	}()

	return m, nil
}

func (m *Multicast) Close() error {
	if m.listen != nil {
		return m.listen.Close()
	}

	return nil
}

func (m *Multicast) initListen(conf *Conf) (*AutoListen, error) {
	if !conf.Multicast.Listen {
		return nil, nil
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
	weSentThis, err := m.listen.IsListening(msg.Src.String())
	if err != nil {
		return err
	}

	if weSentThis {
		return nil
	}

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
