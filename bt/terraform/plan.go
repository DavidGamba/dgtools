package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/mattn/go-isatty"
)

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("plan", "")
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("ignore-cache", false, opt.Description("ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.StringSlice("target", 1, 99)
	opt.SetCommandFn(planRun)

	wss, err := validWorkspaces(cfg)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	destroy := opt.Value("destroy").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	ignoreCache := opt.Value("ignore-cache").(bool)
	varFiles := opt.Value("var-file").([]string)
	targets := opt.Value("target").([]string)
	ws := opt.Value("ws").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg)

	ws, err = getWorkspace(cfg, ws, varFiles)
	if err != nil {
		return err
	}

	defaultVarFiles, err := getDefaultVarFiles(cfg)
	if err != nil {
		return err
	}

	varFiles, err = AddVarFileIfWorkspaceSelected(cfg, ws, varFiles)
	if err != nil {
		return err
	}

	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}

	files, modified, err := fsmodtime.Target(os.DirFS("."),
		[]string{planFile},
		append(append([]string{".tf.init"}, defaultVarFiles...), varFiles...))
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", ".tf.init")
	}
	if !ignoreCache && !modified {
		Logger.Printf("no changes: skipping plan\n")
		return nil
	}
	Logger.Printf("modified: %v\n", files)

	cmd := []string{"terraform", "plan", "-out", planFile}
	for _, v := range defaultVarFiles {
		cmd = append(cmd, "-var-file", v)
	}
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
	}
	if destroy {
		cmd = append(cmd, "-destroy")
	}
	if detailedExitcode {
		cmd = append(cmd, "-detailed-exitcode")
	}
	for _, t := range targets {
		cmd = append(cmd, "-target", t)
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)

	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log()
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		// exit code 2 with detailed-exitcode means changes found
		var eerr *exec.ExitError
		if detailedExitcode && errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			Logger.Printf("plan has changes\n")
			return eerr
		}
		os.Remove(planFile)
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
