package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func stateCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("state", "")

	stateListCMD(ctx, opt)
	statePullCMD(ctx, opt)
	statePushCMD(ctx, opt)
	stateMVCMD(ctx, opt)
	stateRMCMD(ctx, opt)
	stateShowCMD(ctx, opt)
	return opt
}

func stateListCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("list", "")
	opt.SetCommandFn(stateListRun)
	return opt
}

func stateListRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "list"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func statePushCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("push", "")
	opt.SetCommandFn(statePushRun)
	opt.HelpSynopsisArg("<state_file>", "State file to push")
	return opt
}

func statePushRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <state_file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	stateFile := args[0]
	args = slices.Delete(args, 0, 1)

	cfg := config.ConfigFromContext(ctx)
	invalidateCacheContext(ctx, true)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "push", stateFile}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func statePullCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("pull", "")
	opt.SetCommandFn(statePullRun)
	return opt
}

func statePullRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "pull"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func stateMVCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("mv", "")
	opt.SetCommandFn(stateMVRun)
	return opt
}

func stateMVRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	invalidateCacheContext(ctx, true)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "mv"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func stateRMCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("rm", "")
	opt.SetCommandFn(stateRMRun)
	return opt
}

func stateRMRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	invalidateCacheContext(ctx, true)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "rm"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func stateShowCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("show", "")
	opt.SetCommandFn(stateShowRun)
	return opt
}

func stateShowRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "state", "show"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
