package udpmux

import (
	"crypto/cipher"
	"net"
	"time"
)

type Instance struct {
	Network   string
	Addr      *net.UDPAddr
	Remote    *net.UDPAddr
	Channels  []*net.UDPAddr
	Sockets   []*net.UDPConn
	MuxSocket *net.UDPConn
	AEAD      cipher.AEAD
	Tag       []byte

	*ClientInstance
}

type ClientInstance struct {
	Interval time.Duration
	Output   *RotateWriter
}

func (i *Instance) IsServer() bool {
	return i.ClientInstance == nil
}
