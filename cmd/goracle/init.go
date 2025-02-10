package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sean9999/gork"
	"github.com/sean9999/hermeti"
	"github.com/spf13/afero"
)

var privOut io.Writer = nil
var pubOut io.Writer = nil
var confOut afero.File = nil

func resolvePath(s string) (string, error) {
	bits := strings.Split(s, afero.FilePathSeparator)
	if bits[0] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		bits[0] = home
		s = filepath.Join(bits...)
	}
	return s, nil
}

// gork init initializes a [gork.Principal]
func (cmd *Exe) Init(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	//	the default output stream is whatever the env says
	privOut = env.OutStream
	pubOut = env.OutStream

	fset := flag.NewFlagSet("dirfinder", flag.ContinueOnError)
	fset.BoolFunc("o", "directory in which to save keys", func(s string) error {

		if s == "" || s == "true" {
			path, err := resolvePath(DefaultDirectory)
			if err != nil {
				return err
			}
			s = path
		}

		stat, err := os.Stat(s)
		if err != nil {
			return fmt.Errorf("%w: %s", err, s)
		}
		if !stat.IsDir() {
			return errors.New("not a dir")
		}
		//	the overridden output stream is whatever directory the command-line flag says
		privOut, err = os.Create(filepath.Join(s, "priv.pem"))
		if err != nil {
			return err
		}
		pubOut, err = os.Create(filepath.Join(s, "pub.pem"))
		if err != nil {
			return err
		}
		confOut, err = os.Create(filepath.Join(s, "conf.json"))
		if err != nil {
			return err
		}
		return nil
	})
	err := fset.Parse(args)
	if err != nil {
		fmt.Fprintln(env.ErrStream, err)
		return args, err
	}

	prov := gork.FileBasedConfigProvider{
		Fs:   afero.NewOsFs(),
		Name: "conf.json",
	}
	p := gork.Principal{}
	if confOut == nil {
		p = gork.NewPrincipal(env.Randomness, nil, nil)
	} else {
		p = gork.NewPrincipal(env.Randomness, nil, prov)
	}

	cmd.Self = &p

	privPem, err := p.MarshalPEM()
	if err != nil {
		fmt.Fprintln(env.ErrStream, err)
		return args, err
	}

	pubPem, err := p.AsPeer().MarshalPEM()
	if err != nil {
		fmt.Fprintln(env.ErrStream, err)
		return args, err
	}

	// err = exe.Self.Save(exe.ConfigFile)
	// if err != nil {
	// 	return nil, pear.Errorf("couldn't save: %w", err)
	// }

	//p.Save()

	fmt.Fprintf(privOut, "%s\n", privPem)
	fmt.Fprintf(pubOut, "%s\n", pubPem)

	if fpriv, ok := privOut.(afero.File); ok {
		os.Chmod(fpriv.Name(), 0400)
	}

	return args, nil
}
