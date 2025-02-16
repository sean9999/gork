package main

import (
	"context"
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
	self := cli.Obj().Self
	confIn := self.Export()
	err = self.LoadConfig(confIn)
	assert.NoError(t, err)
	err = self.Save(nil)
	assert.NoError(t, err)
	confOut := self.Export()
	foo, ok := confOut.Props.Get("foo")
	assert.Equal(t, "bar", foo)
	assert.True(t, ok)

}
