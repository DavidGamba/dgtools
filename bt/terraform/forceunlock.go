package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func forceUnlockCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("force-unlock", "")
	opt.SetCommandFn(forceUnlockRun)
	opt.HelpSynopsisArg("<lock-id>", "Lock ID")

	return opt
}

func forceUnlockRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <lock-id>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	lockID := args[0]
	args = slices.Delete(args, 0, 1)

	cfg := config.ConfigFromContext(ctx)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "force-unlock", "-force", lockID}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
