package main

import (
	"encoding/pem"
	"errors"
	"fmt"
	"net"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/gork"
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
	s := spool{
		conn, inbox, outbox, errs,
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
			inbox <- env
		}
	}()
	return s
}

func processAssertion(exe state, inEnv Envelope, errs chan error, outbox chan Envelope) {
	if inEnv.Message.Subject != "ASSERTION" {
		errs <- errors.New("bad subject")
	}

	valid := inEnv.Message.Valid()
	verified := inEnv.Message.Verify()
	if !valid || !verified {
		errs <- errors.New("not valid or verified")
	}

	me := exe.self
	conf := exe.conf

	//	extract peer and ensure it comes with an address
	peerKey := inEnv.Message.Sender
	peer := gork.NewPeer(peerKey.Bytes())
	peer.Properties.Set("addr", inEnv.SenderAddress.String())

	//	add peer
	err := me.AddPeer(peer)
	if err != nil {
		errs <- err
	}

	//	save to config
	err = me.Save(conf)
	if err != nil {
		errs <- err
	}

	//	let's send an ACK back
	msg := me.Compose([]byte("I friended you."), nil, peer)
	msg.Headers.Set("you_can_contact_me_at", exe.localAddr.String())
	msg.Sign(exe.environment.Randomness, me)

	outEnv := Envelope{
		Message:          msg,
		SenderAddress:    exe.localAddr,
		RecipientAddress: inEnv.SenderAddress,
	}

	fmt.Println(outEnv)

	outbox <- outEnv
}
