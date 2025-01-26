package main

import (
	"context"
	"encoding/pem"
	"io"
	"testing"

	"github.com/sean9999/go-delphi"
	"github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {

	check := assert.New(t)

	cli := SetupTestCLI(t)
	cli.Env.Args = []string{"goracle", "assert", "--priv", "../../testdata/george.pem"}

	ctx := context.TODO()
	cli.Run(ctx)

	r, err := cli.OutStream()
	check.NoError(err)

	pemBytes, err := io.ReadAll(r)
	check.NoError(err)
	check.Contains(string(pemBytes), "ASSERTION")

	pemBlock, _ := pem.Decode(pemBytes)
	msg := &delphi.Message{}

	err = msg.FromPEM(*pemBlock)
	check.NoError(err)

	check.Equal("ASSERTION", msg.Subject)
	msg.Valid()

	me := cli.Obj().Self

	digest, err := msg.Digest()
	check.NoError(err)

	valid := me.Verify(msg.Sender, digest, msg.Signature())

	check.True(valid)

}
