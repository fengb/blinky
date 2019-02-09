package main

import (
	"fmt"
	"github.com/mingrammer/commonregex"
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

type connMeta struct {
	Conn      *net.UDPConn
	Iface     net.Interface
	IpStrings []string
}

type AutoListen struct {
	Addr       *net.UDPAddr
	C          chan ReadMsg
	metas      map[string]*connMeta
	ipStrings  map[string]bool
	updateDone chan struct{}
}

func NewAutoListen(addr *net.UDPAddr) (*AutoListen, error) {
	a := AutoListen{
		addr,
		make(chan ReadMsg),
		make(map[string]*connMeta),
		make(map[string]bool),
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
	closeAll := make([]func() error, 0, len(a.metas))
	for _, meta := range a.metas {
		closeAll = append(closeAll, meta.Conn.Close)
	}

	err := Parallel(closeAll...)
	close(a.C)
	close(a.updateDone)
	return err
}

func (a *AutoListen) IsListening(raw string) (bool, error) {
	ips := commonregex.IPs(raw)
	if len(ips) == 0 {
		return false, fmt.Errorf("'%s' does not contain recognizable IP", raw)
	}
	if len(ips) > 1 {
		return false, fmt.Errorf("'%s' matched too many IPs %s", raw, ips)
	}
	_, isListening := a.ipStrings[ips[0]]
	return isListening, nil
}

func (a *AutoListen) connect(iface net.Interface) error {
	_, isConnected := a.metas[iface.Name]
	if isConnected {
		return nil
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}

	if iface.Flags&net.FlagMulticast != net.FlagMulticast {
		return nil
	}

	conn, err := net.ListenMulticastUDP("udp", &iface, a.Addr)
	if err != nil {
		return err
	}

	meta := connMeta{conn, iface, a.extractIpStrings(addrs)}
	a.metas[iface.Name] = &meta
	for _, ip := range meta.IpStrings {
		a.ipStrings[ip] = true
	}
	log.Println("Connected to", iface.Name, meta.IpStrings)

	go func() {
		defer a.disconnect(iface)

		conn.SetReadBuffer(maxPacketSize)
		for {
			packet := make([]byte, maxPacketSize)

			n, src, err := conn.ReadFromUDP(packet)
			if err != nil {
				_, isConnected := a.metas[iface.Name]
				if !isConnected {
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
	meta, isConnected := a.metas[iface.Name]
	if !isConnected {
		return nil
	}

	err := meta.Conn.Close()
	delete(a.metas, iface.Name)
	for _, ip := range meta.IpStrings {
		delete(a.ipStrings, ip)
	}

	return err
}

func (a *AutoListen) extractIpStrings(addrs []net.Addr) []string {
	strings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		strings = append(strings, commonregex.IPs(addr.String())...)
	}
	return strings
}
