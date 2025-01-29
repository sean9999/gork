package main

import (
	"context"
	"fmt"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
)

// gork init initializes a [gork.Principal]
func (cmd *Exe) Init(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	p := gork.NewPrincipal(env.Randomness, nil, nil)

	cmd.Self = &p

	pem, err := p.MarshalPEM()

	if err != nil {
		fmt.Fprintln(env.ErrStream, err)
		return args, err
	}

	fmt.Fprintf(env.OutStream, "%s\n", pem)
	return args, nil
}
