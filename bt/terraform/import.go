package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func importCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("import", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(importRun)

	return opt
}

func importRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	i := invalidatePlan{}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	return varFileCMDRun(i, cfg.TFProfile[profile].BinaryName, "import")(ctx, opt, args)
}
