package main

import (
	"context"

	"github.com/sean9999/hermeti"
	"github.com/sean9999/pear"
)

func (exe *Exe) Save(ctx context.Context, env hermeti.Env, args []string) ([]string, error) {
	args, err := exe.ensureSelf(ctx, env, args)
	if err != nil {
		return nil, pear.Errorf("couldn't save: %w", err)
	}

	err = exe.Self.Save(exe.ConfigFile)
	if err != nil {
		return nil, pear.Errorf("couldn't save: %w", err)
	}

	// j, err := json.MarshalIndent(exe.Self.Config, "", "\t")
	// if err != nil {
	// 	return nil, pear.Errorf("coulidn't marshal config: %w", err)
	// }

	// fmt.Fprintf(env.OutStream, "%s\n", j)

	return args, nil
}
