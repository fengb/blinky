package main

import (
	"log"
	"net"
	"time"
)

const maxPacketSize = 512

type Multicast struct {
	conf          *Conf
	aes           *Aes
	listen        *net.UDPConn
	ping          *time.Ticker
	pingLocalAddr net.Addr
}

func NewMulticast(conf *Conf) (Actor, error) {
	m := Multicast{conf: &Conf{}}
	err := m.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

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

	ping, err := m.updatePing(conf)
	if err != nil {
		return err
	}
	if m.ping != ping && m.ping != nil {
		m.ping.Stop()
	}

	aes, err := NewAes(conf.Multicast.Secret)
	if err != nil {
		return err
	}

	m.aes = aes
	m.listen = listen
	m.ping = ping
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

func (m *Multicast) updatePing(conf *Conf) (*time.Ticker, error) {
	if conf.Multicast.Ping == 0 {
		return nil, nil
	}

	if conf.Multicast.Ping == m.conf.Multicast.Ping &&
		conf.Multicast.Addr == m.conf.Multicast.Addr {
		return m.ping, nil
	}

	addr, err := net.ResolveUDPAddr("udp", conf.Multicast.Addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	m.pingLocalAddr = conn.LocalAddr()

	ticker := time.NewTicker(conf.Multicast.Ping)
	go func() {
		defer conn.Close()
		for _ = range ticker.C {
			msg, err := m.aes.Encrypt([]byte("Send Packet"))
			if err != nil {
				log.Println(err)
			}
			conn.Write(msg)
		}
	}()

	return ticker, nil
}

func (m *Multicast) handleListen(src *net.UDPAddr, packet []byte) {
	if m.pingLocalAddr == nil || m.pingLocalAddr.String() == src.String() {
		// We sent this. Ignore!
		return
	}

	names, _ := net.LookupAddr(src.IP.String())
	msg, err := m.aes.Decrypt(packet)
	if err != nil {
		log.Println(err)
	}
	log.Println("Received from", names, src, string(msg))
}
