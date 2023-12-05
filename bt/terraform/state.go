package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
)

func stateListCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-list", "")
	opt.SetCommandFn(stateListRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func stateListRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	cmd := []string{cfg.Terraform[profile].BinaryName, "state", "list"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func statePushCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-push", "")
	opt.SetCommandFn(statePushRun)
	opt.HelpSynopsisArg("<state_file>", "State file to push")

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

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
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	cmd := []string{cfg.Terraform[profile].BinaryName, "state", "push", stateFile}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func statePullCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-pull", "")
	opt.SetCommandFn(statePullRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func statePullRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	cmd := []string{cfg.Terraform[profile].BinaryName, "state", "pull"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func stateRMCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-rm", "")
	opt.SetCommandFn(stateRMRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func stateRMRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	cmd := []string{cfg.Terraform[profile].BinaryName, "state", "rm"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}

func stateShowCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("state-show", "")
	opt.SetCommandFn(stateShowRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func stateShowRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	cmd := []string{cfg.Terraform[profile].BinaryName, "state", "show"}
	return wsCMDRun(cmd...)(ctx, opt, args)
}
