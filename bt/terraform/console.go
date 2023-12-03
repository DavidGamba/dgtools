package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func consoleCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("console", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(consoleRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func consoleRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	i := noOp{}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg)

	return varFileCMDRun(i, cfg.Terraform.BinaryName, "console")(ctx, opt, args)
}
