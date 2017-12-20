package main

import (
	"log"
	"net"
)

const maxPacketSize = 8192

type ReadMsg struct {
	Packet []byte
	Src    *net.UDPAddr
}

type MulticastListen struct {
	Addr      *net.UDPAddr
	C         chan ReadMsg
	listeners map[string]*net.UDPConn
}

func NewMulticastListen(addr *net.UDPAddr) (*MulticastListen, error) {
	m := MulticastListen{addr, make(chan ReadMsg), make(map[string]*net.UDPConn)}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagMulticast == net.FlagMulticast {
			go m.connect(iface)
		}
	}

	return &m, nil
}

func (m *MulticastListen) Close() error {
	closeAll := make([]func() error, 0)
	for _, conn := range m.listeners {
		closeAll = append(closeAll, conn.Close)
	}

	err := Parallel(closeAll...)
	close(m.C)
	return err
}

func (m *MulticastListen) connect(iface net.Interface) error {
	conn, err := net.ListenMulticastUDP("udp", &iface, m.Addr)
	if err != nil {
		return err
	}

	m.listeners[iface.Name] = conn

	defer func() {
		m.listeners[iface.Name].Close()
		m.listeners[iface.Name] = nil
	}()

	log.Println("Connected to", iface.Name)
	conn.SetReadBuffer(maxPacketSize)
	for {
		packet := make([]byte, maxPacketSize)

		n, src, err := conn.ReadFromUDP(packet)
		if err != nil {
			log.Println(err)
			continue
		}

		m.C <- ReadMsg{packet[:n], src}
	}
}

func (m *MulticastListen) disconnect(iface net.Interface) error {
	if m.listeners[iface.Name] == nil {
		return nil
	}

	err := m.listeners[iface.Name].Close()
	m.listeners[iface.Name] = nil

	return err
}
