package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//

func TestInfo(t *testing.T) {

	check := assert.New(t)

	cli := SetupTestCLI(t)
	cli.Env.Args = []string{"goracle", "info", "--priv", "../../testdata/aged-smoke.pem"}

	ctx := context.TODO()
	cli.Run(ctx)

	r, err := cli.OutStream()
	check.NoError(err)

	buf := &strings.Builder{}
	_, err = io.Copy(buf, r)
	check.NoError(err)

	check.Contains(buf.String(), "e4e7cfb70470a569aa0450d708e524bdc1211b8a9fae219ea22f50f4b339c220be522655c322247ca73bfa3995d0d7f41628b4a95550d64d11e42e8a311803a8")

}
