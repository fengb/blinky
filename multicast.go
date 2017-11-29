package main

import (
	"log"
	"net"
	"time"
)

const maxPacketSize = 512

type Multicast struct {
	conf   *Conf
	listen *net.UDPConn
	ping   *time.Ticker
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
	m.listen = listen

	ping, err := m.updatePing(conf)
	if err != nil {
		return err
	}
	if m.ping != ping && m.ping != nil {
		m.ping.Stop()
	}
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

	ticker := time.NewTicker(conf.Multicast.Ping)
	go func() {
		defer conn.Close()
		for _ = range ticker.C {
			conn.Write([]byte("Send Packet"))
		}
	}()

	return ticker, nil
}

func (m *Multicast) handleListen(src *net.UDPAddr, packet []byte) {
	log.Println("Received from", src, string(packet))
}
