package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func initCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("init", "")
	opt.Bool("dry-run", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the init"), opt.Alias("ic"))
	opt.SetCommandFn(InitRun)
	return opt
}

func InitRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	color := opt.Value("color").(string)
	ignoreCache := opt.Value("ignore-cache").(bool)
	automation := opt.Value("tf-in-automation").(bool)
	ws := opt.Value("ws").(string)

	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)
	os.Setenv("CONFIG_ROOT", cfg.ConfigRoot)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	if cfg.TFProfile[cfg.Profile(profile)].Workspaces.Enabled {
		if automation && !workspaceSelected(cfg.Config.DefaultTerraformProfile, profile) {
			if ws == "" {
				return fmt.Errorf("running in workspace mode in automation but no workspace selected or --ws given")
			}
		}
	}

	lockFile := ".terraform.lock.hcl"
	initFile := ".tf.init"
	files, modified, err := fsmodtime.Target(os.DirFS(dir), []string{initFile}, []string{lockFile})
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", lockFile)
	}
	if !ignoreCache && !modified {
		Logger.Printf("no changes: skipping init\n")
		return nil
	}
	if len(files) > 0 {
		Logger.Printf("modified: %v\n", files)
	} else {
		Logger.Printf("missing target: %v\n", initFile)
	}

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "init"}

	for _, bvars := range cfg.TFProfile[cfg.Profile(profile)].Init.BackendConfig {
		b := strings.ReplaceAll(bvars, "~", "$HOME")
		bb, err := fsmodtime.ExpandEnv([]string{b}, nil)
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		// TODO: Consider re-introducing validation
		// if _, err := os.Stat(bb[0]); err == nil {
		// }
		cmd = append(cmd, "-backend-config", bb[0])
	}
	if color == "never" || (color == "auto" && !isatty.IsTerminal(os.Stdout.Fd())) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)
	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		if automation {
			dataDir = fmt.Sprintf("%s-%s", dataDir, ws)
		}
	}
	Logger.Printf("export %s\n", dataDir)
	err = run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun).Run()
	if err != nil {
		os.Remove(filepath.Join(dir, ".tf.lock"))
		return fmt.Errorf("failed to run: %w", err)
	}

	if dryRun {
		return nil
	}

	os.Remove(filepath.Join(dir, ".tf.lock"))
	initFilePath := filepath.Join(dir, initFile)
	fh, err := os.Create(initFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()
	Logger.Printf("Create %s\n", initFilePath)

	return nil
}
