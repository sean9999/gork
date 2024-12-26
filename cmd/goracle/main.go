package main

import (
	"os"

	"github.com/sean9999/go-flargs"
	"github.com/sean9999/gork/cmd/goracle/subcommand"
)

func main() {

	env := flargs.NewCLIEnvironment("/")
	env.Arguments = os.Args[1:]

	flargset, err := subcommand.ParseGlobals(env)

	if err != nil {
		complain("could not parse globals", 5, nil, env.ErrorStream)
	}

	//	consume the first argument, or "info"
	subcmd := "info"
	if len(flargset.Remainders) > 0 {
		subcmd = flargset.Remainders[0]
		flargset.Remainders = flargset.Remainders[1:]
	}

	switch subcmd {

	case "info":
		subcommand.Info(env, flargset)
	case "init":
		subcommand.Init(env, flargset)
	// case "assert":
	// 	err = subcommand.Assert(env, *globals)
	// case "echo":
	// 	err = subcommand.Echo(env)
	// case "sign":
	// 	err = subcommand.Sign(env, globals)
	// case "verify", "add-peer":
	// 	err = subcommand.Verify(env, globals)
	// case "peers":
	// 	err = subcommand.Peers(env, globals)
	// case "encrypt":
	// 	err = subcommand.Encrypt(env, globals, remainingArgs)
	// case "decrypt":
	// 	err = subcommand.Decrypt(env, globals, remainingArgs)

	default:
		complain("unsupported subcommand", 3, nil, env.ErrorStream)
	}

	// if err != nil {
	// 	complain(fmt.Sprintf("subcommand %s", remainingArgs[0]), 7, err, env.ErrorStream)
	// }

}
