package main

import (
	"context"
	"testing"

	"github.com/sean9999/gork"
	"github.com/stretchr/testify/assert"
)

// func setupCli(t testing.TB) *hermeti.CLI[*Exe] {
// 	t.Helper()
// 	//	a test CLI uses a testing environment
// 	env := hermeti.TestEnv()
// 	env.Randomness = streamOf(1)
// 	//	instatiate the object that represents our CLI
// 	cmd := new(Exe)

// 	//	wrap it in hermeti.CLI
// 	cli := &hermeti.CLI[*Exe]{
// 		Env: env,
// 		Cmd: cmd,
// 	}
// 	return cli
// }

func TestInit(t *testing.T) {

	ctx := context.Background()
	cli := SetupTestCLI(t)
	cli.Env.Args = []string{"prog", "init"}
	cli.Run(ctx)

	pemFile, err := cli.OutStream()
	assert.NoError(t, err)

	pubkey1 := cli.Obj().Self.PublicKey()

	prince := new(gork.Principal)
	prince.FromPem(pemFile)

	pubkey2 := prince.PublicKey()

	pubkey1.Equal(pubkey2)

}
