package cmd

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"sync"
)

const RXBUFLEN = 2048

type Tunnel struct {
	local   netip.AddrPort
	remote  netip.AddrPort
	tunnel  int
	urx     chan []byte
	trx     chan []byte
	tun     *os.File
	conn    *net.UDPConn
	verbose bool
	wg      sync.WaitGroup
}

func NewTunnel(tunnel int, localAddr string, localPort int, remoteAddr string, remotePort int, verbose bool) (*Tunnel, error) {
	local, err := netip.ParseAddr(localAddr)
	if err != nil {
		return nil, err
	}
	remote, err := netip.ParseAddr(remoteAddr)
	if err != nil {
		return nil, err
	}
	t := Tunnel{
		local:   netip.AddrPortFrom(local, uint16(localPort)),
		remote:  netip.AddrPortFrom(remote, uint16(remotePort)),
		tunnel:  tunnel,
		urx:     make(chan []byte),
		trx:     make(chan []byte),
		verbose: verbose,
	}
	return &t, nil
}

func (t *Tunnel) ListenUdp() error {
	udpAddr := net.UDPAddrFromAddrPort(t.local)
	conn, err := net.ListenUDP("udp4", udpAddr)
	t.conn = conn
	if err != nil {
		return fmt.Errorf("ListenUDP failed: %v", err)
	}
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer close(t.urx)
		remoteAddr := net.UDPAddrFromAddrPort(t.remote)
		for {
			buf := make([]byte, RXBUFLEN)
			count, source, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("ReadFromUDP failed: %v", err)
				return
			}
			if count > 0 {
				if source.IP.Equal(remoteAddr.IP) && source.Port == remoteAddr.Port {
					t.urx <- buf
				} else {
					log.Printf("source: %+v\n", source)
					log.Printf("remote: %+v\n", remoteAddr)
					log.Printf("ignoring packet from unexpected source: %+v", source)
				}
			}
		}
	}()
	return nil
}

func (t *Tunnel) ListenTun() error {
	device := fmt.Sprintf("/dev/tun%d", t.tunnel)
	tun, err := os.OpenFile(device, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("tunnel device open failed: %v", err)
	}
	t.tun = tun
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer close(t.trx)
		for {
			t.wg.Add(1)
			buf := make([]byte, RXBUFLEN)
			_, err := t.tun.Read(buf)
			if err != nil {
				log.Printf("ListenTun: Read failed: %v", err)
				return
			}
			t.trx <- buf
		}
	}()
	return nil
}

func (t *Tunnel) Run() error {
	err := t.ListenUdp()
	if err != nil {
		return err
	}
	err = t.ListenTun()
	if err != nil {
		return err
	}
	go func() {
		defer t.tun.Close()
		defer t.conn.Close()
		defer log.Println("awaiting listener shutdown...")
		for {
			select {
			case packet, ok := <-t.urx:
				if ok {
					if t.verbose {
						log.Printf("URX: %+v\n", packet)
					}
					count, err := t.tun.Write(packet)
					if err != nil {
						log.Printf("tunnel write failed: %v", err)
						return
					} else if count != len(packet) {
						log.Printf("tunnel write count mismatch: len=%d sent=%d", len(packet), count)
					}
				} else {
					log.Println("UDP channel closed")
					return
				}

			case packet, ok := <-t.trx:
				if ok {
					if t.verbose {
						log.Printf("TRX: %+v\n", packet)
					}
					count, err := t.conn.WriteToUDPAddrPort(packet, t.remote)
					if err != nil {
						log.Printf("UDP write failed: %v", err)
						return
					} else if count != len(packet) {
						log.Printf("UDP write count mismatch: len=%d sent=%d", len(packet), count)
					}
				} else {
					log.Println("tunnel channel closed")
					return
				}
			}
		}
	}()
	t.wg.Wait()
	log.Println("shutdown complete")
	return nil
}
