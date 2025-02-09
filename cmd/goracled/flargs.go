package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

func flargs(args []string) (port uint, conf string, priv string, err error) {

	flagset := flag.NewFlagSet("flagset", flag.PanicOnError)
	flagset.UintVar(&port, "port", 5656, "specify port")
	flagset.StringVar(&conf, "config", "config.json", "config file")
	flagset.StringVar(&priv, "priv", "key.pem", "private key")
	err = flagset.Parse(args)

	fmt.Println(port, conf, priv, err)

	return port, conf, priv, err

}

func setup(conf string) (state, error) {
	s := state{}
	f, err := afero.NewMemMapFs().Open(conf)
	s.conf = f
	return s, err
}

func initialize(filesystem afero.Fs, env hermeti.Env) (state, error) {
	s := state{}
	port, confName, privName, err := flargs(env.Args)
	if err != nil {
		return s, err
	}
	s.port = port

	conf, err := filesystem.OpenFile(confName, os.O_RDWR|os.O_TRUNC, 0640)

	if err != nil {
		return s, err
	}
	s.conf = conf
	s.environment = env

	priv, err := filesystem.Open(privName)
	if err != nil {
		return s, err
	}

	p := new(gork.Principal)
	err = p.FromPem(priv)
	p.WithRand(env.Randomness)
	p.WithConfigFile(conf)
	s.self = p

	return s, err
}
