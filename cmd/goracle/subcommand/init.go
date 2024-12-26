package subcommand

import (
	"crypto/rand"
	"fmt"

	"github.com/sean9999/go-flargs"
	"github.com/sean9999/gork"
)

func Init(env *flargs.Environment, flargset *FlargSet) {

	complain := NewComplainer(env.ErrorStream)

	kv := map[string]string{
		"io.gork/version": "v0.0.1",
	}
	prince := gork.NewPrincipal(rand.Reader, kv)
	pemBytes, err := prince.MarshalPEM()
	if err != nil {
		complain(err, 1)
	}
	fmt.Fprintf(env.OutputStream, "%s\n", pemBytes)
}
