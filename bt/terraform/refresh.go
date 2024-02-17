package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func refreshCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("refresh", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(refreshRun)
	return opt
}

func refreshRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	i := invalidatePlan{}

	cfg := config.ConfigFromContext(ctx)
	invalidateCacheContext(ctx, true)
	LogConfig(cfg, profile)

	return varFileCMDRun(i, cfg.TFProfile[cfg.Profile(profile)].BinaryName, "refresh")(ctx, opt, args)
}
