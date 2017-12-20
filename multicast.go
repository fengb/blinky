package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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
	listen        *MulticastListen
	sendCache     []byte
	send          *net.UDPConn
	sendTimer     *time.Timer
}

func NewMulticast(conf *Conf, snapshotState *SnapshotState) (*Multicast, error) {
	m := Multicast{
		conf:          &Conf{},
		snapshotState: snapshotState,
		sendTimer:     time.NewTimer(1 * time.Hour),
	}

	err := m.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		update := func(snapshot *Snapshot) {
			m.sendCache, err = m.encode(snapshot)
			if err != nil {
				log.Println(err)
			}

			m.sendTimer.Stop()
			m.sendTimer.Reset(1 * time.Nanosecond)
		}

		if snapshotState.Local() != nil {
			update(snapshotState.Local())
		}

		for snapshot := range snapshotState.SubLocal() {
			update(snapshot)
		}
	}()

	go func() {
		for _ = range m.sendTimer.C {
			if m.send == nil {
				// Bad state
				log.Fatal("sendConn is nil... why")
			}

			m.send.Write(m.sendCache)
			m.sendTimer.Reset(sendInterval)
		}
	}()

	return &m, nil
}

func (m *Multicast) UpdateConf(conf *Conf) error {
	var (
		aes    *Aes
		listen *MulticastListen
		send   *net.UDPConn
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
		func() (err error) {
			send, err = m.updateSend(conf)
			return err
		},
	)

	if err != nil {
		closeAll(listen, send)
		return err
	}

	if send == nil {
		m.sendTimer.Stop()
	}

	if m.listen != listen {
		closeAll(m.listen)
		m.listen = listen
	}
	if m.send != send {
		closeAll(m.send)
		m.send = send
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

func (m *Multicast) updateListen(conf *Conf) (*MulticastListen, error) {
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

	listener, err := NewMulticastListen(addr)
	if err != nil {
		return nil, err
	}

	go func() {
		for msg := range listener.C {
			if m.send != nil && m.send.LocalAddr().String() == msg.Src.String() {
				// We sent this. Ignore!
				continue
			}

			err = m.snapshotState.UpdateNetworkLink(msg.Src, msg.Packet, m.decode)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return listener, nil
}

func (m *Multicast) updateSend(conf *Conf) (*net.UDPConn, error) {
	if !conf.Multicast.Send {
		return nil, nil
	}

	if conf.Multicast.Send == m.conf.Multicast.Send &&
		conf.Multicast.Addr == m.conf.Multicast.Addr {
		return m.send, nil
	}

	addr, err := net.ResolveUDPAddr("udp", conf.Multicast.Addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
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
