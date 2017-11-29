package main

import (
	"log"
	"net"
)

const maxPacketSize = 512

type Multicast struct {
	conf   *Conf
	listen *net.UDPConn
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
	if m.listen != nil && (!conf.Multicast.Listen || conf.Multicast.Addr != m.conf.Multicast.Addr) {
		err := m.listen.Close()
		if err != nil {
			return err
		}
		m.listen = nil
	}

	if conf.Multicast.Listen && m.listen == nil {
		addr, err := net.ResolveUDPAddr("udp", conf.Multicast.Addr)
		if err != nil {
			return err
		}

		conn, err := net.ListenMulticastUDP("udp", nil, addr)
		if err != nil {
			return err
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
				m.handle(src, packet[:n])
			}
		}()

		m.listen = conn
	}

	m.conf = conf

	return nil
}

func (m *Multicast) handle(src *net.UDPAddr, packet []byte) {
	log.Println("Packet received", packet)
}
