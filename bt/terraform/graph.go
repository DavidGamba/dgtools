package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func graphCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("graph", "")
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("plan", false, opt.Description("Use the latest plan file"))
	opt.String("T", "png", opt.Description("Set output format. For example: -T png"))
	opt.String("filename", "", opt.Description("Set output filename"))
	opt.SetCommandFn(graphRun)

	return opt
}

func graphRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	ws := opt.Value("ws").(string)
	varFiles := opt.Value("var-file").([]string)
	plan := opt.Value("plan").(bool)
	format := opt.Value("T").(string)
	filename := opt.Value("filename").(string)

	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}
	ws, err = getWorkspace(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}

	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)
	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "graph"}
	if plan {
		cmd = append(cmd, "-plan", planFile)
	}
	cmd = append(cmd, args...)
	ri := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir)
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	out, err := ri.STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	fmt.Printf("%s\n", string(out))
	if opt.Called("T") {
		if filename == "" {
			if ws == "" {
				filename = fmt.Sprintf("graph.%s", format)
			} else {
				filename = fmt.Sprintf("graph-%s.%s", ws, format)
			}
		}
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		err = run.CMDCtx(ctx, "dot", "-T", format).Log().In(out).Run(f, os.Stderr)
		if err != nil {
			return fmt.Errorf("failed to generate graph: %w", err)
		}
		Logger.Printf("graph saved to: %s\n", filename)
	}

	return nil
}
