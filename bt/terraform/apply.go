package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func applyCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("apply", "")
	opt.SetCommandFn(applyRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func applyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	ws := opt.Value("ws").(string)
	profile := opt.Value("profile").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	if cfg.TFProfile[profile].Workspaces.Enabled {
		if !workspaceSelected() {
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
	files, modified, err := fsmodtime.Target(os.DirFS("."), []string{applyFile}, []string{planFile})
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", applyFile)
	}
	if !modified {
		Logger.Printf("no changes: skipping apply\n")
		return nil
	}
	Logger.Printf("modified: %v\n", files)

	cmd := []string{cfg.TFProfile[profile].BinaryName, "apply"}
	cmd = append(cmd, "-input", planFile)
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)
	dataDir := fmt.Sprintf("TF_DATA_DIR=.terraform-%s", profile)
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log().Env(dataDir)
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		os.Remove(planFile)
		return fmt.Errorf("failed to run: %w", err)
	}

	fh, err := os.Create(applyFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()
	return nil
}
