package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func validateCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("validate", "")
	opt.SetCommandFn(validateRun)

	return opt
}

func validateRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[cfg.Profile(profile)])

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "validate"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
