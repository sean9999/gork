package main

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClobberProps(t *testing.T) {

	ctx := context.Background()
	cli := SetupTestCLI(t)

	fd, err := cli.Env.Filesystem.Open("../../testdata/kv.txt")
	assert.NoError(t, err)

	cli.Env.InStream = fd
	cli.Env.Args = []string{"goracle", "props", "--priv", "../../testdata/young-dew.pem", "--config", "../../testdata/young-dew.config.json"}
	cli.Run(ctx)

	conf := cli.Obj().Self.Export()
	fyle, err := io.ReadAll(conf)

	assert.NoError(t, err)
	assert.Contains(t, string(fyle), "foo")
	assert.Contains(t, string(fyle), "bar")

}
