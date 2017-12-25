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

type AutoListen struct {
	Addr       *net.UDPAddr
	C          chan ReadMsg
	conns      map[string]*net.UDPConn
	updateDone chan struct{}
}

func NewAutoListen(addr *net.UDPAddr) (*AutoListen, error) {
	a := AutoListen{
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
		go a.connect(iface)
	}

	go func() {
		updates := make(chan netlink.LinkUpdate)
		netlink.LinkSubscribe(updates, a.updateDone)
		for update := range updates {
			iface, err := net.InterfaceByName(update.Attrs().Name)
			if err != nil {
				log.Println(err)
				continue
			}

			if update.Flags&unix.IFF_RUNNING == unix.IFF_RUNNING {
				a.connect(*iface)
			} else {
				a.disconnect(*iface)
			}
		}
	}()

	return &a, nil
}

func (a *AutoListen) Close() error {
	closeAll := make([]func() error, 0, len(a.conns))
	for _, conn := range a.conns {
		closeAll = append(closeAll, conn.Close)
	}

	err := Parallel(closeAll...)
	close(a.C)
	close(a.updateDone)
	return err
}

func (a *AutoListen) isConnected(iface net.Interface) bool {
	return a.conns[iface.Name] != nil
}

func (a *AutoListen) connect(iface net.Interface) error {
	if a.isConnected(iface) {
		return nil
	}

	if iface.Flags&net.FlagMulticast != net.FlagMulticast {
		return nil
	}

	conn, err := net.ListenMulticastUDP("udp", &iface, a.Addr)
	if err != nil {
		return err
	}

	a.conns[iface.Name] = conn
	log.Println("Connected to", iface.Name)

	go func() {
		defer a.disconnect(iface)

		conn.SetReadBuffer(maxPacketSize)
		for {
			packet := make([]byte, maxPacketSize)

			n, src, err := conn.ReadFromUDP(packet)
			if err != nil {
				if !a.isConnected(iface) {
					return
				}

				log.Println(err)
				continue
			}

			a.C <- ReadMsg{packet[:n], src}
		}
	}()

	return nil
}

func (a *AutoListen) disconnect(iface net.Interface) error {
	if !a.isConnected(iface) {
		return nil
	}

	err := a.conns[iface.Name].Close()
	a.conns[iface.Name] = nil

	return err
}
