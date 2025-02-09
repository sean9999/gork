package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
	"github.com/spf13/afero"
)

var (
	DefaultDirectory = filepath.Join("~", ".gork")
)

// an CLIError is an error with an exit code
type CLIError struct {
	Msg      string
	ExitCode int
	Child    error
}

func (o *CLIError) Error() string {
	if o.Child == nil {
		return o.Msg
	} else {
		return fmt.Sprintf("%s: %s", o.Msg, o.Child)
	}
}

func (o *CLIError) Wrap(child error) {
	o.Child = child
}

func (o *CLIError) Unrawp() error {
	return o.Child
}

func complain(msg string, exitCode int, child error, stream io.Writer) {
	err := &CLIError{msg, exitCode, child}
	fmt.Fprintln(stream, err)
	os.Exit(exitCode)
}

// Exe is the execution of a command, including state
type Exe struct {
	Verbosity  uint
	Self       *gork.Principal
	ConfigFile afero.File
}

func (e *Exe) State() *Exe {
	return e
}

type subcommand func(context.Context, hermeti.Env, []string) ([]string, error)

// Run sets the whole thing in motion and contains the entire execution lifecycle
func (exe *Exe) Run(env hermeti.Env) {

	args := env.Args[1:]
	ctx := context.Background()

	args, err := exe.bootstrap(ctx, env, args)
	if err != nil {
		panic(err)
	}

	//	if no subcommand is specified, "info" is implied
	subcmd := "info"

	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}

	//	modify this as needed
	subcommands := map[string]subcommand{
		"info":   exe.Info,
		"init":   exe.Init,
		"save":   exe.Save,
		"assert": exe.Assert,
		"add":    exe.Add,
		"export": exe.Export,
	}

	fn, exists := subcommands[subcmd]

	if !exists {
		err := pear.Errorf("unsupported command: %q", subcmd)
		panic(err)
	}

	_, err = fn(ctx, env, args)
	if err != nil {
		fmt.Fprintln(env.ErrStream, err)
	}

}

// parse global values early in execution
func (cmd *Exe) bootstrap(_ context.Context, _ hermeti.Env, args []string) ([]string, error) {

	gset := flag.NewFlagSet("global", flag.ContinueOnError)
	verbosity := gset.Uint("verbosity", 0, "verbosity level")
	gset.Parse(args)

	cmd.Verbosity = *verbosity
	args = gset.Args()

	return args, nil

}

// ensureSelf ensures the presence of a gork.Principal by checking for --priv and optionally --config
func (cmd *Exe) ensureSelf(_ context.Context, env hermeti.Env, args []string) ([]string, error) {

	if cmd.Self != nil {
		return args, nil
	}

	fset := flag.NewFlagSet("selfer", flag.ContinueOnError)
	conf := new(string)
	priv := new(string)
	fset.StringVar(conf, "config", "~/.gork/config.json", "config file location")
	fset.StringVar(priv, "priv", "~/.gork/priv.pem", "private key location")
	fset.Parse(args)

	//	the lack of a well-formed pem file is fatal
	pemFile, err := env.Filesystem.Open(*priv)
	if err != nil {
		return args, pear.Errorf("could not find pem file: %w", err)
	}

	pemBytes, err := io.ReadAll(pemFile)
	if err != nil {
		return args, pear.Errorf("could not read pem file: %w", err)
	}

	p := gork.NewPrincipal(env.Randomness, nil, nil)

	err = p.UnmarshalPEM(pemBytes)
	if err != nil {
		return args, pear.Errorf("could not create principal: %w", err)
	}
	cmd.Self = &p

	//	the lack of a config file is not an error
	confFile, err := env.Filesystem.OpenFile(*conf, os.O_RDWR|os.O_CREATE, 0644)
	if err == nil {
		err = p.WithConfigFile(confFile)
		if err != nil {
			return nil, err
		}
		cmd.ConfigFile = confFile
	}

	return fset.Args(), nil

}
