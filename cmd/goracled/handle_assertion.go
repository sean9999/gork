package main

import (
	"errors"
	"fmt"

	"github.com/sean9999/gork"
)

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
	msg.Subject = "ORACLE MESSAGE"
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
