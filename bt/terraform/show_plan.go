package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func showPlanCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("show-plan", "Show the cached terraform plan")
	opt.Bool("dry-run", false)
	opt.SetCommandFn(showPlanRun)
	return opt
}

func showPlanRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	ws := opt.Value("ws").(string)
	color := opt.Value("color").(string)

	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	if cfg.TFProfile[cfg.Profile(profile)].Workspaces.Enabled {
		if !workspaceSelected(cfg.Config.DefaultTerraformProfile, profile) {
			if ws == "" {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
		}
	}

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "show"}
	cmd = append(cmd, args...)
	// possitional arg goes at the end
	if ws == "" {
		cmd = append(cmd, ".tf.plan")
	} else {
		cmd = append(cmd, fmt.Sprintf(".tf.plan-%s", ws))
	}
	if color == "never" || (color == "auto" && !isatty.IsTerminal(os.Stdout.Fd())) {
		cmd = append(cmd, "-no-color")
	}
	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun)
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
