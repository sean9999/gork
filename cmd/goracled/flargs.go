package main

import (
	"flag"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

func flargs(args []string) (port uint64, conf string, priv string, err error) {
	flagset := flag.NewFlagSet("flagset", flag.PanicOnError)
	flagset.Uint64Var(&port, "port", 0, "specify port")
	flagset.StringVar(&conf, "config", "config.json", "config file")
	flagset.StringVar(&priv, "priv", "key.pem", "private key")
	err = flagset.Parse(args)
	return port, conf, priv, err
}

func initialize(filesystem afero.Fs, env hermeti.Env) (state, error) {
	s := state{}
	port, confName, privName, err := flargs(env.Args)
	if err != nil {
		return s, err
	}
	s.port = uint16(port)
	prov := gork.FileBasedConfigProvider{
		Fs:   env.Filesystem,
		Name: confName,
	}
	s.conf = prov
	s.environment = env
	priv, err := filesystem.Open(privName)
	if err != nil {
		return s, err
	}

	p := gork.NewPrincipal(env.Randomness, nil, prov)
	err = p.FromPem(priv)
	if err != nil {
		return s, err
	}
	err = p.WithConfigProvider(prov)
	// p.Props = gork.NewKV()
	// p.WithRand(env.Randomness)
	// err = p.WithConfigProvider(prov)
	if err != nil {
		return s, err
	}
	s.self = &p
	return s, err
}
