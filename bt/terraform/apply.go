package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func applyCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("apply", "")
	opt.Bool("dry-run", false)
	opt.SetCommandFn(applyRun)
	return opt
}

func applyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	ws := opt.Value("ws").(string)
	profile := opt.Value("profile").(string)

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

	applyFile := ""
	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
		applyFile = ".tf.apply"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
		applyFile = fmt.Sprintf(".tf.apply-%s", ws)
	}
	files, modified, err := fsmodtime.Target(os.DirFS(dir), []string{applyFile}, []string{planFile})
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", applyFile)
	}
	if !modified {
		Logger.Printf("no changes: skipping apply\n")
		return nil
	}
	Logger.Printf("modified: %v\n", files)

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "apply"}
	cmd = append(cmd, "-input", planFile)
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)
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
		os.Remove(filepath.Join(dir, planFile))
		return fmt.Errorf("failed to run: %w", err)
	}

	if dryRun {
		return nil
	}

	fh, err := os.Create(filepath.Join(dir, applyFile))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()
	Logger.Printf("Create %s\n", filepath.Join(dir, applyFile))

	return nil
}
