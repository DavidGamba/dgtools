package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func outputCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("output", "")
	opt.SetCommandFn(outputRun)

	return opt
}

func outputRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	cmd := []string{cfg.TFProfile[profile].BinaryName, "output"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
