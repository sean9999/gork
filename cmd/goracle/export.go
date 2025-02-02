package main

import (
	"context"
	"fmt"

	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
)

func (cmd *Exe) Export(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, pear.Errorf("%w: %w. Could not ensure self.", ErrAssert, err)
	}

	msg := cmd.Self.Bytes()

	fmt.Fprintf(env.OutStream, "%X\n", msg)

	return args, nil

}
