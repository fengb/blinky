package main

import (
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"log"
	"net"
)

const maxPacketSize = 8192

type ReadMsg struct {
	Packet []byte
	Src    *net.UDPAddr
}

type MulticastListen struct {
	Addr           *net.UDPAddr
	C              chan ReadMsg
	listeners      map[string]*net.UDPConn
	linkUpdateDone chan struct{}
}

func NewMulticastListen(addr *net.UDPAddr) (*MulticastListen, error) {
	m := MulticastListen{
		addr,
		make(chan ReadMsg),
		make(map[string]*net.UDPConn),
		make(chan struct{}),
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		go m.connect(iface)
	}

	go func() {
		updates := make(chan netlink.LinkUpdate)
		netlink.LinkSubscribe(updates, m.linkUpdateDone)
		for update := range updates {
			iface, err := net.InterfaceByName(update.Attrs().Name)
			if err != nil {
				log.Println(err)
				continue
			}

			if update.Flags&unix.IFF_RUNNING == unix.IFF_RUNNING {
				m.connect(*iface)
			} else {
				m.disconnect(*iface)
			}
		}
	}()

	return &m, nil
}

func (m *MulticastListen) Close() error {
	var closeAll [](func() error)
	for _, conn := range m.listeners {
		closeAll = append(closeAll, conn.Close)
	}

	err := Parallel(closeAll...)
	close(m.C)
	close(m.linkUpdateDone)
	return err
}

func (m *MulticastListen) connect(iface net.Interface) error {
	if m.listeners[iface.Name] != nil {
		return nil
	}

	if iface.Flags&net.FlagMulticast != net.FlagMulticast {
		return nil
	}

	conn, err := net.ListenMulticastUDP("udp", &iface, m.Addr)
	if err != nil {
		return err
	}

	m.listeners[iface.Name] = conn
	log.Println("Connected to", iface.Name)

	go func() {
		defer m.disconnect(iface)

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
	}()

	return nil
}

func (m *MulticastListen) disconnect(iface net.Interface) error {
	if m.listeners[iface.Name] == nil {
		return nil
	}

	err := m.listeners[iface.Name].Close()
	m.listeners[iface.Name] = nil

	return err
}
