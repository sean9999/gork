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
	cli.Env.Args = []string{"goracle", "info", "--priv", "../../testdata/george.pem"}

	ctx := context.TODO()
	cli.Run(ctx)

	r, err := cli.OutStream()
	check.NoError(err)

	buf := &strings.Builder{}
	_, err = io.Copy(buf, r)
	check.NoError(err)

	check.Contains(buf.String(), "5c2f587f8fa3f324412f1e033e2ecd8eb32d8e3f7703c38f5533eb11f7d9ad38583e7047ea1533d1b503e6fc40f4ea7bb7c5e5501fe442c311e60478776e464f")

}
