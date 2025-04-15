package terraform

import (
	"context"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func testCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("test", "")
	opt.StringSlice("var-file", 1, 1)
	opt.SetCommandFn(testRun)

	return opt
}

func testRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	i := noOp{}

	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	return varFileCMDRun(i, cfg.TFProfile[cfg.Profile(profile)].BinaryName, "test")(ctx, opt, args)
}
