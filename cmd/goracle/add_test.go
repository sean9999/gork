package main

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {

	//	me
	pem := "../../testdata/late-silence.pem"
	conf := "../../testdata/late-silence.config.json"

	//	person to add
	assertion := "../../testdata/aged-smoke.assertion.pem"

	check := assert.New(t)
	cli := SetupTestCLI(t)

	//	pipe assertion to stdin
	fd, err := cli.Env.Filesystem.Open(assertion)
	check.NoError(err)
	io.Copy(cli.Env.InStream.(io.Writer), fd)

	//	launch the CLI
	cli.Env.Args = []string{"goracle", "add", "--priv", pem, "--config", conf}
	ctx := context.TODO()
	cli.Run(ctx)

	//	capture output
	outstream, err := cli.OutStream()
	check.NoError(err)
	result, err := io.ReadAll(outstream)
	check.NoError(err)
	check.Contains(string(result), "96c1e46")
	check.Contains(string(result), "aged-smoke")

}
