package main

import (
	"context"
	"testing"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	//	a test CLI uses a testing environment
	env := hermeti.TestEnv()

	env.Randomness = streamOf(1)

	ctx := context.Background()
	//	instatiate the object that represents our CLI
	cmd := new(Exe)

	//	wrap it in hermeti.CLI
	cli := &hermeti.CLI[*Exe]{
		Env: env,
		Cmd: cmd,
	}

	cli.Env.Args = []string{"prog", "init"}

	cli.Run(ctx)

	pemFile, err := cli.OutStream()
	assert.NoError(t, err)

	pubkey1 := cmd.Self.PublicKey()

	prince := new(gork.Principal)
	prince.FromPem(pemFile)

	pubkey2 := prince.PublicKey()

	pubkey1.Equal(pubkey2)

}
