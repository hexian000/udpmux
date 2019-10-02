package udpmux

import (
	"log"
	"net"
)

func (i *Instance) Run() {
	conn, err := net.ListenUDP(i.Network, i.Addr)
	if err != nil {
		log.Fatalln("ListenUDP:", err)
	}
	defer func() { _ = conn.Close() }()
	i.MuxSocket = conn
	log.Println("listen ", conn.LocalAddr())

	i.Sockets = make([]*net.UDPConn, len(i.Channels))
	remotes := make([]*net.UDPAddr, len(i.Channels))
	for c := 0; c < len(i.Channels); c++ {
		go i.mux(conn, uint16(c), &remotes[c])
	}

	if !i.IsServer() && i.ClientInstance.Interval > 0 {
		go i.keepalive()
	}

	b := make([]byte, maxPacketSize)
	for {
		n, r, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("server:", err)
			continue
		}
		if i.Remote == nil {
			i.Remote = r
		}
		h := &header{}
		if msg := i.unpack(h, b[:n]); msg != nil {
			if h.channel == keepaliveChan {
				if i.IsServer() {
					i.OnEcho(msg)
				} else {
					i.OnKeepAlive(msg)
				}
				continue
			}

			var sendTo *net.UDPAddr
			if i.IsServer() {
				sendTo = i.Channels[h.channel]
			} else {
				sendTo = remotes[h.channel]
			}
			log.Printf("[%d] demux %v -> %v", h.channel, r, sendTo)
			_, _ = i.Sockets[h.channel].WriteToUDP(msg, sendTo)
			continue
		}
		log.Println("server unpack failed.")
	}
}

func (i *Instance) mux(muxConn *net.UDPConn, c uint16, addr **net.UDPAddr) {
	var local *net.UDPAddr
	if i.IsServer() {
		local = nil
	} else {
		local = i.Channels[c]
	}
	conn, err := net.ListenUDP(i.Network, local)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() { _ = conn.Close() }()
	log.Println("channel ", i.Channels[c])
	i.Sockets[c] = conn

	b := make([]byte, maxPacketSize)
	for {
		n, remote, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("ReadFromUDP:", err)
		}
		*addr = remote
		packet := i.pack(&header{reserved: 0, channel: c}, b[:n])
		log.Printf("[%d] mux %v -> %v\n", c, remote, i.Remote)
		_, _ = muxConn.WriteToUDP(packet, i.Remote)
	}
}
