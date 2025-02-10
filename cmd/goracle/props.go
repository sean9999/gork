package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/sean9999/hermeti"
)

// add props to self, overwritting any existing props
func (cmd *Exe) ClobberProps(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {

	args, err := cmd.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(env.InStream)
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			fmt.Fprintf(env.ErrStream, "this seems to be badly formed: %v", kv)
			continue
		}

		props := cmd.Self.Props

		x := *props

		x.Set(kv[0], kv[1])

		fmt.Fprintln(env.OutStream, line)
	}
	cmd.Self.Save(cmd.Self.ConfigProvider)

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(env.ErrStream, err)
	}
	return args, err
}
