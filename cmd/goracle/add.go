package main

import (
	"context"
	"encoding/pem"
	"errors"
	"io"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
)

var ErrAdd = pear.Defer("could not add peer")

func wrap(e error) error {
	return pear.Errorf("%w: %w", ErrAdd, e)
}

func is(e error) bool {
	return e != nil
}

func (cmd *Exe) Add(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if is(err) {
		return args, wrap(err)
	}

	pemBytes, err := io.ReadAll(env.InStream)
	if is(err) {
		return args, wrap(err)
	}

	pemBlock, _ := pem.Decode(pemBytes)
	msg := &delphi.Message{}

	err = msg.FromPEM(*pemBlock)
	if is(err) {
		return args, wrap(err)
	}

	me := cmd.Self
	dig, err := msg.Digest()
	if is(err) {
		return args, wrap(err)
	}

	valid := me.Verify(msg.Sender, dig, msg.Signature())
	if !valid {
		return args, wrap(errors.New("invalid signature"))
	}

	peer := gork.NewPeer(msg.Sender.Bytes())

	err = me.AddPeer(peer)
	if is(err) {
		return args, wrap(err)
	}

	//	output the full config
	_, err = io.Copy(env.OutStream, me.Export())
	if is(err) {
		return args, wrap(err)
	}
	return args, nil
}
