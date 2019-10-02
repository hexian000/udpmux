package udpmux

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

const msgKeepalive uint32 = iota

var seq uint64 = 0

type KeepAliveMsg struct {
	msg       uint32
	seq       uint64
	timestamp int64
}

func (k *KeepAliveMsg) read(b []byte) {
	k.msg = binary.BigEndian.Uint32(b[0:4])
	k.seq = binary.BigEndian.Uint64(b[4:12])
	k.timestamp = int64(binary.BigEndian.Uint64(b[12:20]))
}

func (k *KeepAliveMsg) write(b []byte) {
	binary.BigEndian.PutUint32(b[0:4], k.msg)
	binary.BigEndian.PutUint64(b[4:12], k.seq)
	binary.BigEndian.PutUint64(b[12:20], uint64(k.timestamp))
}

func (k *KeepAliveMsg) len() int { const keepaliveLen = 20; return keepaliveLen }

func (i *Instance) OnEcho(b []byte) {
	packet := i.pack(&header{reserved: 0, channel: keepaliveChan}, b)
	_, _ = i.MuxSocket.WriteToUDP(packet, i.Remote)
}

func (i *Instance) OnKeepAlive(b []byte) {
	msg := &KeepAliveMsg{}
	msg.read(b)
	now := time.Now().UnixNano()
	rtt := (now - msg.timestamp) / int64(time.Millisecond)
	log.Println("RTT:", rtt, "ms")
	if i.Output != nil {
		_, _ = io.WriteString(i.Output, fmt.Sprintf(
			"\"%v\"\t%d\t%d\n",
			time.Unix(0, msg.timestamp).Format(time.RFC3339),
			msg.seq, rtt,
		))
	}
}

func (i *Instance) keepalive() {
	t := time.NewTicker(i.ClientInstance.Interval)
	defer t.Stop()

	msg := &KeepAliveMsg{msg: msgKeepalive}
	b := make([]byte, msg.len())
	for range t.C {
		msg.seq = seq
		msg.timestamp = time.Now().UnixNano()
		msg.write(b)
		packet := i.pack(&header{reserved: 0, channel: keepaliveChan}, b)
		_, _ = i.MuxSocket.WriteToUDP(packet, i.Remote)
		seq++
	}
}
