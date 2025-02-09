package subcommand

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sean9999/go-flargs"
	"github.com/sean9999/gork"
)

var lineBreak = []byte("\n")

// func argument(str string) string {
// 	pattern := regexp.MustCompile(`\w+`)
// 	return pattern.FindString(str)
// }

// func flaag(str string) string {
// 	return fmt.Sprintf("--%s", argument(str))
// }

func configOrDie(f *FlargSet) error {
	fset := flag.NewFlagSet("info", flag.ContinueOnError)
	conf := new(string)
	seed := new(string)
	fset.StringVar(conf, "config", "~/.gork/config.json", "config file location")
	fset.StringVar(seed, "priv", "", "private key location")
	err := fset.Parse(f.Remainders)
	if err != nil {
		return err
	}
	f.Set("config", *conf)
	f.Set("priv", *seed)
	f.Remainders = fset.Args()
	return nil
}

// Info outputs public information about oneself.
func Info(env *flargs.Environment, flargset *FlargSet) {

	complain := NewComplainer(env.ErrorStream)

	err := flargset.Parse(configOrDie)
	if err != nil {
		complain(err, 7)
	}

	priv, ok := flargset.Get("priv")
	if !ok {
		complain(errors.New("no private key"), 6)
	}

	pemFile, err := env.Filesystem.OpenFile(priv.(string), os.O_RDONLY, 0644)
	if err != nil {
		complain(err, 9)
	}

	pemBytes, err := io.ReadAll(pemFile)
	if err != nil {
		complain(err, 11)
	}

	p := new(gork.Principal)

	err = p.UnmarshalPEM(pemBytes)
	if err != nil {
		complain(err, 11)
	}

	fmt.Fprintln(env.OutputStream, p.Art())

	for k, v := range p.Properties.Entries() {
		fmt.Fprintf(env.OutputStream, "%s\t%s\n", k, v)
	}

	fmt.Fprintf(env.OutputStream, "grip:\t%s\n", p.AsPeer().Grip())

	// type outputFormat struct {
	// 	Self  oracle.Peer            `json:"self"`
	// 	Peers map[string]oracle.Peer `json:"peers,omitempty"`
	// }

	// me, err := oracle.From(globals.Config)
	// if err != nil {
	// 	return err
	// }

	// ooo := outputFormat{
	// 	Self:  me.AsPeer(),
	// 	Peers: me.Peers(),
	// }

	// j, _ := json.MarshalIndent(ooo, "", "\t")

	// j, err := me.AsPeer().MarshalJSON()
	// if err != nil {
	// 	return err
	// }

	//env.OutputStream.Write(j)
	//env.OutputStream.Write(lineBreak)

	// conf, ok := flargset.Get("config")
	// if !ok {
	// 	//fmt.Fprintln(env.ErrorStream, "no config set")
	// 	return errors.New("no config")
	// }

	// b, err := env.Filesystem.ReadFile(conf.(string))
	// if err != nil {
	// 	return errors.New("no config set")
	// }

	// fmt.Fprintf(env.OutputStream, "config is %s\n\n", b)

	// if len(me.Peers()) > 0 {
	// 	env.OutputStream.Write(lineBreak)
	// 	env.OutputStream.Write([]byte("peers"))
	// 	env.OutputStream.Write(lineBreak)
	// 	for nick := range me.Peers() {
	// 		env.OutputStream.Write([]byte(nick))
	// 		env.OutputStream.Write(lineBreak)
	// 	}
	// }

}
