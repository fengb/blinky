package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const (
	maxPacketSize = 8192
	sendInterval  = 30 * time.Second
)

type Multicast struct {
	conf          *Conf
	snapshotState *SnapshotState
	aes           *Aes
	listen        *net.UDPConn
	sendCache     []byte
	send          *net.UDPConn
	sendTimer     *time.Timer
}

func NewMulticast(conf *Conf, snapshotState *SnapshotState) (*Multicast, error) {
	m := Multicast{
		conf:          &Conf{},
		snapshotState: snapshotState,
		sendTimer:     time.NewTimer(1 * time.Second),
	}

	err := m.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	m.sendCache, err = m.encode(snapshotState.Local())
	if err != nil {
		return nil, err
	}

	go func() {
		for snapshot := range snapshotState.SubLocal() {
			m.sendCache, err = m.encode(snapshot)
			if err != nil {
				log.Println(err)
			}
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
		listen *net.UDPConn
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
		cleanupConn(listen, send)
		return err
	}

	if send == nil {
		m.sendTimer.Stop()
	}

	if m.listen != listen {
		cleanupConn(m.listen)
		m.listen = listen
	}
	if m.send != send {
		cleanupConn(m.send)
		m.send = send
	}

	m.aes = aes
	m.conf = conf

	return nil
}

func cleanupConn(conns ...*net.UDPConn) {
	for _, conn := range conns {
		if conn == nil {
			continue
		}

		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

func (m *Multicast) updateListen(conf *Conf) (*net.UDPConn, error) {
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

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	log.Println("Listening on", addr)

	go func() {
		conn.SetReadBuffer(maxPacketSize)
		for {
			packet := make([]byte, maxPacketSize)
			n, src, err := conn.ReadFromUDP(packet)
			if err != nil {
				log.Println(err)
				continue
			}

			if m.send != nil && m.send.LocalAddr().String() == src.String() {
				// We sent this. Ignore!
				continue
			}

			err = m.snapshotState.UpdateNetworkLink(src, packet[:n], m.decode)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return conn, nil
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
