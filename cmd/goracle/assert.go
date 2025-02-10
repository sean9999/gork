package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
)

var ErrAssert = errors.New("could not assert")

func (cmd *Exe) Assert(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, pear.Errorf("%w: %w. Could not ensure self.", ErrAssert, err)
	}

	//	by including props in the body
	//	we can ensure the integrity of those too.
	//	props are included has headers, but headers are not used in digest calculation
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

	for pair := msg.Headers.Oldest(); pair != nil; pair = pair.Next() {
		msg.Headers.Set(pair.Key, pair.Value)
	}

	err = msg.Sign(env.Randomness, &cmd.Self)
	if err != nil {
		return nil, pear.Errorf("%w: %w. Could not sign message", ErrAssert, err)
	}

	fmt.Fprintf(env.OutStream, "%s\n", msg)

	return args, nil

}
