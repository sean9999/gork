package main

import (
	"errors"
	"fmt"

	"github.com/sean9999/gork"
)

func processAllBase(exe state, inEnv Envelope, errs chan error, outbox chan Envelope) {
	if inEnv.Message.Subject != "ALL BASE" {
		errs <- errors.New("bad subject")
	}

	if !inEnv.Message.Valid() {
		errs <- errors.New("not valid")
	} else {
		fmt.Println("valid")
	}
	if !inEnv.Message.Verify() {
		errs <- errors.New("not verified")
	} else {
		fmt.Println("verified")
	}

	me := exe.self
	//conf := exe.conf

	//	extract peer and ensure it comes with an address
	peerKey := inEnv.Message.Sender
	peer := gork.NewPeer(peerKey.Bytes())

	sentence := fmt.Sprintf("My nickname is %s and the senders nickname is %s", me.Nickname(), peer.Nickname())

	fmt.Println(sentence)

	//	say hello to all my friends
	for _, p := range me.Peers {
		addr, err := p.Address()
		if err != nil {
			errs <- fmt.Errorf("%w for %s", err, p.Nickname())
		}
		msg := me.Compose([]byte(sentence), nil, p)
		msg.Subject = "HELLO"
		env := Envelope{
			SenderAddress:    exe.localAddr,
			RecipientAddress: addr,
			Message:          msg,
		}
		outbox <- env

	}

}
