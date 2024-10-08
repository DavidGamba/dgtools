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
	opt.SetCommandFn(initRun)
	return opt
}

func initRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	color := opt.Value("color").(string)

	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)
	os.Setenv("CONFIG_ROOT", cfg.ConfigRoot)

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
	Logger.Printf("export %s\n", dataDir)
	err := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun).Run()
	if err != nil {
		os.Remove(filepath.Join(dir, ".tf.lock"))
		return fmt.Errorf("failed to run: %w", err)
	}

	if dryRun {
		return nil
	}

	os.Remove(filepath.Join(dir, ".tf.lock"))
	initFile := filepath.Join(dir, ".tf.init")
	fh, err := os.Create(initFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()
	Logger.Printf("Create %s\n", initFile)

	return nil
}
