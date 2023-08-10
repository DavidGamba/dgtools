package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func statePushCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-push", "")
	opt.SetCommandFn(statePushRun)
	opt.HelpSynopsisArg("<state_file>", "State file to push")

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func statePushRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <state_file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	stateFile := args[0]
	args = slices.Delete(args, 0, 1)

	cmd := []string{"terraform", "state", "push", stateFile}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func statePullCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-pull", "")
	opt.SetCommandFn(statePullRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func statePullRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	cmd := []string{"terraform", "state", "pull"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
