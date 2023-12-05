package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func consoleCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("console", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(consoleRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func consoleRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	i := noOp{}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	return varFileCMDRun(i, cfg.TFProfile[profile].BinaryName, "console")(ctx, opt, args)
}
