package main

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {

	check := assert.New(t)
	cli := SetupTestCLI(t)

	//	pipe assertion to stdin
	fd, err := cli.Env.Filesystem.Open("../../testdata/george.assertion.pem")
	check.NoError(err)
	io.Copy(cli.Env.InStream.(io.Writer), fd)

	//	launch the CLI
	cli.Env.Args = []string{"goracle", "add", "--priv", "../../testdata/john.pem", "--config", "../../testdata/john.config.json"}
	ctx := context.TODO()
	cli.Run(ctx)

	//	capture output
	outstream, err := cli.OutStream()
	check.NoError(err)
	result, err := io.ReadAll(outstream)
	check.NoError(err)
	check.Contains(string(result), "98351ddf")
	check.Contains(string(result), "shy-pine")

}
