package udpmux

import "encoding/binary"

const (
	maxPacketSize = 65536
	keepaliveChan = uint16(0xFFFF)
)

type header struct {
	reserved uint16
	channel  uint16
}

func (h *header) read(b []byte) {
	h.reserved = binary.BigEndian.Uint16(b[0:2])
	h.channel = binary.BigEndian.Uint16(b[2:4])
}

func (h *header) write(b []byte) {
	binary.BigEndian.PutUint16(b[0:2], h.reserved)
	binary.BigEndian.PutUint16(b[2:4], h.channel)
}

func (h *header) len() int { const headerLen = 4; return headerLen }
