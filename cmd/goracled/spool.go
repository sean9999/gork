package main

import (
	"net"

	"github.com/sean9999/go-delphi"
)

type blob []byte

type spool struct {
	conn   net.PacketConn
	inbox  chan delphi.Message
	outbox chan delphi.Message
}

func NewSpool(conn net.PacketConn) spool {
	inbox := make(chan delphi.Message)
	outbox := make(chan delphi.Message)
	return spool{
		conn, inbox, outbox,
	}
}
