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
	sendInterval  = 1 * time.Second
)

type receiveCacheData struct {
	packet   []byte
	snapshot *Snapshot
}

type Multicast struct {
	conf         *Conf
	pac          *Pac
	aes          *Aes
	sendCache    []byte
	receiveCache map[string]receiveCacheData
	listen       *net.UDPConn
	send         *net.UDPConn
	sendTimer    *time.Timer
}

func NewMulticast(conf *Conf, pac *Pac) (Actor, error) {
	m := Multicast{
		conf:         &Conf{},
		pac:          pac,
		receiveCache: make(map[string]receiveCacheData),
		sendTimer:    time.NewTimer(sendInterval),
	}
	err := m.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range m.sendTimer.C {
			if m.send == nil {
				// Bad state
				log.Fatal("sendConn is nil... why")
			}

			if m.sendCache == nil {
				snapshot, err := m.pac.GetSnapshot()
				if err != nil {
					log.Println(err)
				}
				packet, err := m.encode(snapshot)
				if err != nil {
					log.Println(err)
				}
				m.sendCache = packet
			}

			m.send.Write(m.sendCache)
			m.sendTimer.Reset(sendInterval)
		}
	}()

	return &m, nil
}

func (m *Multicast) UpdateConf(conf *Conf) error {
	listen, err := m.updateListen(conf)
	if err != nil {
		return err
	}

	if m.listen != listen && m.listen != nil {
		err = m.listen.Close()
		if err != nil {
			log.Println(err)
		}
	}

	send, err := m.updateSend(conf)
	if err != nil {
		return err
	}
	if send == nil {
		m.sendTimer.Stop()
	}
	if m.send != send && m.send != nil {
		err = m.send.Close()
		if err != nil {
			log.Println(err)
		}
	}

	aes, err := NewAes(conf.Multicast.Secret)
	if err != nil {
		return err
	}

	m.aes = aes
	m.listen = listen
	m.send = send
	m.conf = conf

	return nil
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
			}
			m.handleListen(src, packet[:n])
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

func (m *Multicast) handleListen(src *net.UDPAddr, packet []byte) {
	if m.send != nil && m.send.LocalAddr().String() == src.String() {
		// We sent this. Ignore!
		return
	}

	ipString := src.IP.String()
	cache, ok := m.receiveCache[ipString]
	if ok && bytes.Equal(packet, cache.packet) {
		// Already cached!
		return
	}

	snapshot, err := m.decode(packet)
	if err != nil {
		log.Println(err)
		return
	}

	m.receiveCache[ipString] = receiveCacheData{packet, snapshot}

	names, _ := net.LookupAddr(ipString)
	log.Println("Update received from", names, src, snapshot)
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
