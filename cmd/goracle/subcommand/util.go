package subcommand

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sean9999/go-delphi"
	"github.com/sean9999/go-flargs"
	stablemap "github.com/sean9999/go-stable-map"
	"github.com/sean9999/gork"
)

func NewComplainer(stream io.Writer) func(error, int) {
	return func(err error, exitcode int) {
		fmt.Fprintln(stream, err)
		os.Exit(exitcode)
	}
}

type QQQ struct {
	Arg  string
	Flag string
}

func (q QQQ) Type() string {
	if q.Arg != "" {
		return "arg"
	}
	if q.Flag != "" {
		return "flag"
	}
	return "empty"
}

type FlargSet struct {
	*stablemap.StableMap[string, any]
	Remainders []string
}

func NewFlargSet() *FlargSet {
	smap := stablemap.New[string, any]()
	remainders := []string{}
	return &FlargSet{smap, remainders}
}

func (f *FlargSet) Parse(fn func(*FlargSet) error) error {
	if f == nil {
		return fmt.Errorf("%w: nil flargset", ErrOracle)
	}
	if fn == nil {
		return fmt.Errorf("%w: nil function in flargset.Parse", ErrOracle)
	}
	return fn(f)
}

var ErrOracle = errors.New("oracle")

func wrap(err error) error {
	if errors.Is(err, ErrOracle) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrOracle, err)
}

var ErrNotImplemented = fmt.Errorf("%w: not implemented", ErrOracle)

// an object containing all flags that should be global
type ParamSet struct {
	Format  string
	Config  gork.Config
	Privkey delphi.Key
}

func normalizeHeredoc(inText string) string {
	r := inText
	r = strings.ReplaceAll(r, "\n", " ")
	r = strings.ReplaceAll(r, "\t", " ")
	r = strings.ReplaceAll(r, "   ", " ")
	r = strings.ReplaceAll(r, "  ", " ")
	r = strings.TrimSpace(r)
	return r
}

func looksLikeHexPubkey(s string) bool {
	//	@todo: make this robust
	return len(s) == 128
}

func looksLikeNickname(s string) bool {
	//	@todo: make this robust too
	return (len(s) > 3 && len(s) < 64)
}

func ParseGlobals(env *flargs.Environment) (*FlargSet, error) {

	flargset := NewFlargSet()

	args := env.Arguments

	fset := flag.NewFlagSet("globals", flag.ContinueOnError)

	fset.Func("format", "format to use (pem, ion)", func(s string) error {
		var err error
		switch s {
		case "ion", "pem":
			flargset.Set("format", s)
		default:
			err = flargs.NewFlargError(flargs.ExitCodeGenericError, ErrOracle)
		}
		return err
	})

	fset.Func("conf", "config file", func(s string) error {

		f, err := env.Filesystem.Open(s)
		if err != nil {
			return fmt.Errorf("could not find config file %q. %w", s, err)
		}

		conf := new(gork.Config)
		_, err = io.Copy(conf, f)
		return fmt.Errorf("could not load config from %q. %w", s, err)

	})

	err := fset.Parse(args)
	if err != nil {
		return nil, wrap(err)
	}

	flargset.Remainders = fset.Args()

	return flargset, nil

}
