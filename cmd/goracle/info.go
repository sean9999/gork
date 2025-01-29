package main

import (
	"context"
	"fmt"

	"github.com/sean9999/hermeti"
)

func (cmd *Exe) Info(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(env.OutStream, cmd.Self.AsPeer().Nickname())
	fmt.Fprintf(env.OutStream, "grip:\t%s\n", cmd.Self.AsPeer().Grip())
	fmt.Fprintf(env.OutStream, "pubkey:\t%x\n\n", cmd.Self.PublicKey().Bytes())
	fmt.Fprintf(env.OutStream, "%s\n", cmd.Self.Art())
	return args, err
}
