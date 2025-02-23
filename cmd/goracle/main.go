package main

import (
	"context"

	"github.com/sean9999/hermeti"
)

func main() {

	//	a real CLI uses a real environment
	env := hermeti.RealEnv()
	ctx := context.Background()

	//	capture panics in a pretty stack trace
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		pear.NicePanic(env.ErrStream)
	// 	}
	// }()

	//	instatiate the object that represents our CLI
	cmd := new(Exe)

	//	wrap it in hermeti.CLI
	cli := &hermeti.CLI[*Exe]{
		Env: env,
		Cmd: cmd,
	}

	cli.Run(ctx)

}
