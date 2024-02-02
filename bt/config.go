package main

import (
	"context"
	"fmt"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func configCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("config", "Show bt config file details")
	opt.SetCommandFn(configRun)

	return opt
}

func configRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cfg := config.ConfigFromContext(ctx)
	fmt.Printf("%s", cfg)
	return nil
}
