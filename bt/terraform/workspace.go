package terraform

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func workspaceCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("workspace", "")

	workspaceListCMD(ctx, opt)
	workspaceShowCMD(ctx, opt)
	workspaceSelectCMD(ctx, opt)
	workspaceDeleteCMD(ctx, opt)
	workspaceNewCMD(ctx, opt)

	return opt
}

func workspaceListCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("list", "")
	opt.SetCommandFn(workspaceListRun)

	return opt
}

func workspaceListRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "workspace", "list"}
	return workspaceFn(cmd...)(ctx, opt, args)
}

func workspaceShowCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("show", "")
	opt.SetCommandFn(workspaceShowRun)

	return opt
}

func workspaceShowRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "workspace", "show"}
	return workspaceFn(cmd...)(ctx, opt, args)
}

// When switching to the default workspace, remove the environment file so that we are not in workspace mode
func workspaceSelectCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("select", "")
	opt.SetCommandFn(workspaceSelectRun)
	return opt
}

// TODO: Add autocompletion for workspaces
// TODO: Allow using both the arg and the --ws flag
func workspaceSelectRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	color := opt.Value("color").(string)
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <workspace-name>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	wsName := args[0]
	args = slices.Delete(args, 0, 1)

	cfg := config.ConfigFromContext(ctx)
	LogConfig(cfg, profile)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "workspace", "select", wsName}

	if color == "never" || (color == "auto" && !isatty.IsTerminal(os.Stdout.Fd())) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)
	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)
	err := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	// When switching to the default workspace, remove the environment file so that we are not in workspace mode
	if wsName == "default" {
		dd := getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile))
		os.Remove(fmt.Sprintf("%s/environment", dd))
	}

	return nil
}

func workspaceDeleteCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("delete", "")
	opt.SetCommandFn(workspaceDeleteRun)
	return opt
}

// TODO: Allow using both the arg and the --ws flag
func workspaceDeleteRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "workspace", "delete"}
	return workspaceFn(cmd...)(ctx, opt, args)
}

func workspaceNewCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("new", "")
	opt.SetCommandFn(workspaceNewRun)

	return opt
}

func workspaceNewRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "workspace", "new"}
	return workspaceFn(cmd...)(ctx, opt, args)
}

func workspaceFn(cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		profile := opt.Value("profile").(string)
		color := opt.Value("color").(string)

		cfg := config.ConfigFromContext(ctx)
		LogConfig(cfg, profile)

		if color == "never" || (color == "auto" && !isatty.IsTerminal(os.Stdout.Fd())) {
			cmd = append(cmd, "-no-color")
		}
		cmd = append(cmd, args...)
		dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
		Logger.Printf("export %s\n", dataDir)
		err := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Run()
		if err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}
		return nil
	}
}
