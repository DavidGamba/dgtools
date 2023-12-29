package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func taintCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("taint", "")
	opt.SetCommandFn(taintRun)
	opt.HelpSynopsisArg("<address>", "Address")
	return opt
}

func taintRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <address>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	address := args[0]
	args = slices.Delete(args, 0, 1)

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[cfg.Profile(profile)])

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "taint", address}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func untaintCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("untaint", "")
	opt.SetCommandFn(untaintRun)
	opt.HelpSynopsisArg("<address>", "Address")
	return opt
}

func untaintRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <address>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	address := args[0]
	args = slices.Delete(args, 0, 1)

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[cfg.Profile(profile)])

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "untaint", address}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
