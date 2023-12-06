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
		Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

		ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, profile, ws)
		if err != nil {
			return err
		}

		if cfg.TFProfile[profile].Workspaces.Enabled {
			if !workspaceSelected(cfg.Config.DefaultTerraformProfile, profile) {
				if ws == "" {
					return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
				}
			}
		}

		if !isatty.IsTerminal(os.Stdout.Fd()) {
			cmd = append(cmd, "-no-color")
		}
		cmd = append(cmd, args...)
		dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, profile))
		Logger.Printf("export %s\n", dataDir)
		ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log().Env(dataDir)
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
}
