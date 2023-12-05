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

type VarFileCMDer interface {
	// Function that adds elements to the command based on the workspace
	cmdFunction(ws string) []string

	// Function that runs if the command errored
	errorFunction(ws string)

	// Function that runs if the command succeeded
	successFunction(ws string)
}

type invalidatePlan struct{}

func (fn invalidatePlan) cmdFunction(ws string) []string {
	return []string{}
}

func (fn invalidatePlan) errorFunction(ws string) {
	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}
	os.Remove(planFile)
}

func (fn invalidatePlan) successFunction(ws string) {
	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}
	os.Remove(planFile)
}

type noOp struct{}

func (fn noOp) cmdFunction(ws string) []string {
	return []string{}
}

func (fn noOp) errorFunction(ws string) {}

func (fn noOp) successFunction(ws string) {}

func varFileCMDRun(fn VarFileCMDer, cmd ...string) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		profile := opt.Value("profile").(string)
		varFiles := opt.Value("var-file").([]string)
		ws := opt.Value("ws").(string)
		ws, err := updateWSIfSelected(ws)
		if err != nil {
			return err
		}

		cfg := config.ConfigFromContext(ctx)
		Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

		ws, err = getWorkspace(cfg, profile, ws, varFiles)
		if err != nil {
			return err
		}

		defaultVarFiles, err := getDefaultVarFiles(cfg, profile)
		if err != nil {
			return err
		}

		varFiles, err = AddVarFileIfWorkspaceSelected(cfg, profile, ws, varFiles)
		if err != nil {
			return err
		}

		for _, v := range defaultVarFiles {
			cmd = append(cmd, "-var-file", v)
		}
		for _, v := range varFiles {
			cmd = append(cmd, "-var-file", v)
		}
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			cmd = append(cmd, "-no-color")
		}
		cmd = append(cmd, fn.cmdFunction(ws)...)
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
			fn.errorFunction(ws)
			return fmt.Errorf("failed to run: %w", err)
		}
		fn.successFunction(ws)
		return nil
	}
}
