package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func providersCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("providers", "")
	opt.SetCommandFn(providersRun)

	lock := opt.NewCommand("lock", "")
	lock.SetCommandFn(providersLockRun)
	lock.StringSlice("platform", 1, 99, opt.Description("Target platform"), opt.ArgName("os_arch"))

	return opt
}

func providersRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "providers"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func providersLockRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	platforms := opt.Value("platform").([]string)
	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "providers", "lock"}
	for _, p := range append(platforms, cfg.TFProfile[cfg.Profile(profile)].Platforms...) {
		cmd = append(cmd, "-platform="+p)
	}
	err := wsCMDRun(cmd...)(ctx, opt, args)
	if err != nil {
		return err
	}

	lockFile := filepath.Join(dir, ".tf.lock")
	fh, err := os.Create(lockFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()
	Logger.Printf("Create %s\n", lockFile)
	return nil
}
