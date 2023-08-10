package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func outputCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("output", "")
	opt.SetCommandFn(outputRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func outputRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cmd := []string{"terraform", "output"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
