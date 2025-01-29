package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
)

var ErrAssert = pear.Defer("could not assert")

func (cmd *Exe) Assert(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, pear.Errorf("%w: %w. Could not ensure self.", ErrAssert, err)
	}

	body := struct {
		Msg   string   `json:"msg"`
		Props *gork.KV `json:"props"`
	}{
		"i assert that I am me",
		cmd.Self.Props,
	}

	bodyBytes, jerr := json.Marshal(body)
	if jerr != nil {
		return nil, err
	}

	msg := delphi.NewMessage(env.Randomness, bodyBytes)
	msg.Sender = cmd.Self.PublicKey()
	msg.Subject = "ASSERTION"

	// for k, v := range cmd.Self.Props.Entries() {
	// 	msg.Headers.Set(k, []byte(v))
	// }

	err = msg.Sign(env.Randomness, cmd.Self)
	if err != nil {
		return nil, pear.Errorf("%w: %w. Could not sign message", ErrAssert, err)
	}

	fmt.Fprintf(env.OutStream, "%s\n", msg)

	return args, nil

}
