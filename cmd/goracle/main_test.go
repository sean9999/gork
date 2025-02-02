package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
	"github.com/spf13/afero"
)

// fake randomness. A stream of whatever number you want
type streamOf byte

func (s streamOf) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(s)
	}
	return len(p), nil
}

func SetupTestCLI(t testing.TB) *hermeti.CLI[*Exe] {
	t.Helper()

	//	a test CLI uses a testing environment
	env := hermeti.TestEnv()

	env.Randomness = streamOf(5)

	dir := "../../testdata"

	//	add testdata/** to filesystem
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}
	m := afero.NewMemMapFs()
	for _, e := range entries {
		if !e.IsDir() {
			f, err := os.Open(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Error(err)
				continue
			}
			g, err := m.Create(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Error(err)
			}
			i, err := io.Copy(g, f)
			if err != nil {
				t.Error(err)
			}
			if i == 0 {
				t.Error("zero bytes written")
			}
			f.Close()
		}
	}
	env.Filesystem = m

	//	instatiate the object that represents our CLI
	cmd := new(Exe)

	//	wrap it in hermeti.CLI
	cli := &hermeti.CLI[*Exe]{
		Env: env,
		Cmd: cmd,
	}

	return cli
}

func TestMain(t *testing.T) {

	cli := SetupTestCLI(t)

	//	capture panics in a pretty stack trace
	defer func() {
		if r := recover(); r != nil {
			pear.NicePanic(cli.Env.ErrStream)
		}
	}()

	ctx := context.Background()
	cli.Run(ctx)

}
