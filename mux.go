package udpmux

import (
	"crypto/rand"
	"log"
)

func (i *Instance) unpack(h *header, b []byte) []byte {
	nonceSize := i.AEAD.NonceSize()
	headerSize := h.len()
	if len(b) < (nonceSize + headerSize + i.AEAD.Overhead()) {
		log.Println("short packet")
		return nil
	}
	out, err := i.AEAD.Open([]byte{}, b[:nonceSize], b[nonceSize:], nil)
	if err != nil {
		log.Println("AEAD open failed")
		return nil
	}
	h.read(out[:headerSize])
	return out[headerSize:]
}

func (i *Instance) pack(h *header, b []byte) []byte {
	nonceSize := i.AEAD.NonceSize()
	headerSize := h.len()
	out := make([]byte, len(b)+nonceSize+headerSize+i.AEAD.Overhead())
	nonce := out[:nonceSize]
	_, _ = rand.Read(nonce)
	h.write(out[nonceSize:][:headerSize])
	copy(out[nonceSize+headerSize:], b)
	i.AEAD.Seal(out[nonceSize:][:0], nonce, out[nonceSize:][:headerSize+len(b)], nil)
	return out
}
