package main

import (
	"encoding/pem"
	"fmt"
	"net"

	"github.com/sean9999/go-delphi"
)

const bufSize = 1024

type blob []byte

type Envelope struct {
	Message          *delphi.Message `json:"message"`
	SenderAddress    net.Addr        `json:"sender_addr"`
	RecipientAddress net.Addr        `json:"recipient_addr"`
}
type spool struct {
	conn   net.PacketConn
	inbox  chan Envelope
	outbox chan Envelope
	errors chan error
	logs   chan string
}

func (s spool) Consume(b []byte) (*delphi.Message, error) {
	msg := new(delphi.Message)
	err := msg.UnmarshalBinary(b)
	return msg, err
}

func (s spool) Send(msg delphi.Message, addr net.Addr) error {
	msgAsBytes, err := msg.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = s.conn.WriteTo(msgAsBytes, addr)
	return err
}

type spoolError struct {
	err          error
	bytesWritten int
	addr         net.Addr
}

func (c spoolError) Error() string {
	return c.err.Error()
}

func NewSpool(conn net.PacketConn) spool {

	inbox := make(chan Envelope)
	outbox := make(chan Envelope)
	errs := make(chan error)
	logs := make(chan string, 256)
	s := spool{
		conn, inbox, outbox, errs, logs,
	}

	go func() {
		for {
			//	read in messages and spool them to inbox channel.
			//	anything not well-formed as a delphi.Message is spooled to errors channel.
			buf := make([]byte, bufSize)
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				errs <- spoolError{err, n, addr}
				continue
			}
			msg := new(delphi.Message)
			pemblock, _ := pem.Decode(buf[:n])
			err = msg.FromPEM(*pemblock)
			if err != nil {
				errs <- spoolError{err, n, addr}
				continue
			}
			env := Envelope{
				Message:          msg,
				SenderAddress:    addr,
				RecipientAddress: conn.LocalAddr(),
			}
			logmsg := fmt.Sprintf("got message %q from %q", msg.Subject, addr)
			logs <- logmsg
			inbox <- env
		}
	}()
	return s
}
