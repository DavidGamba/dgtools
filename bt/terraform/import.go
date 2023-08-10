package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func importCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("import", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(importRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func importRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	i := invalidatePlan{}
	return varFileCMDRun(i, "terraform", "import")(ctx, opt, args)
}
