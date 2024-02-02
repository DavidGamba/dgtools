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

func wsCMDRun(cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		profile := opt.Value("profile").(string)
		ws := opt.Value("ws").(string)

		cfg := config.ConfigFromContext(ctx)
		Logger.Printf("cfg: %s\n", cfg.TFProfile[cfg.Profile(profile)])

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

		planFile := ""
		if ws == "" {
			planFile = ".tf.plan"
		} else {
			planFile = fmt.Sprintf(".tf.plan-%s", ws)
		}

		if !isatty.IsTerminal(os.Stdout.Fd()) {
			cmd = append(cmd, "-no-color")
		}
		cmd = append(cmd, args...)
		dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, profile))
		Logger.Printf("export %s\n", dataDir)
		ri := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir)
		if ws != "" {
			wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
			Logger.Printf("export %s\n", wsEnv)
			ri.Env(wsEnv)
		}
		err = ri.Run()
		if err != nil {
			if invalidateCacheFromContext(ctx) {
				os.Remove(planFile)
			}
			return fmt.Errorf("failed to run: %w", err)
		}
		if invalidateCacheFromContext(ctx) {
			os.Remove(planFile)
		}
		return nil
	}
}
