package stack

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/go-getoptions"
)

func ConfigCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("config", "Show stacks config details")
	opt.SetCommandFn(ConfigRun)
	return opt
}

func ConfigRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cfg := config.ConfigFromContext(ctx)

	for _, c := range cfg.Component {
		Logger.Printf("component: %s\n", c)
	}

	for _, s := range cfg.Stack {
		Logger.Printf("stack: %s\n", s)
		for _, c := range s.Components {
			Logger.Printf("%s\n", c)
		}
	}

	return nil
}
