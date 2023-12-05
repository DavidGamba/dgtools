package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

func buildCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("build", "Wraps init, plan and apply into a single operation with a cache")
	opt.SetCommandFn(buildRun)
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.Bool("no-checks", false, opt.Description("Do not run pre-apply checks"), opt.Alias("nc"))
	opt.StringSlice("target", 1, 99)
	opt.StringSlice("replace", 1, 99)
	opt.Bool("apply", false, opt.Description("Apply Terraform plan"))
	opt.Bool("show", false, opt.Description("Show Terraform plan"))

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func buildRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	apply := opt.Value("apply").(bool)
	show := opt.Value("show").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	ws := opt.Value("ws").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.Terraform[profile])

	if cfg.Terraform[profile].Workspaces.Enabled {
		if !workspaceSelected() {
			if ws == "" {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
		}
	}

	initFn := func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		// TODO: Add logic to only run when files have been modified
		if _, err := os.Stat(".tf.init"); os.IsNotExist(err) {
			return initRun(ctx, opt, args)
		}
		return nil
	}

	tm := dag.NewTaskMap()
	tm.Add("init", initFn)
	tm.Add("plan", planRun)
	if cfg.Terraform[profile].PreApplyChecks.Enabled {
		tm.Add("checks", checksRun)
	}
	if apply {
		tm.Add("apply", applyRun)
	}
	if show {
		tm.Add("show", showPlanRun)
	}

	g := dag.NewGraph("build")
	g.TaskDependensOn(tm.Get("plan"), tm.Get("init"))
	if cfg.Terraform[profile].PreApplyChecks.Enabled {
		g.TaskDependensOn(tm.Get("checks"), tm.Get("plan"))
	}

	if show {
		g.TaskDependensOn(tm.Get("show"), tm.Get("plan"))
	}
	if apply {
		g.TaskDependensOn(tm.Get("apply"), tm.Get("plan"))
		if cfg.Terraform[profile].PreApplyChecks.Enabled {
			g.TaskDependensOn(tm.Get("apply"), tm.Get("checks"))
		}
	}
	err = g.Validate(tm)
	if err != nil {
		return fmt.Errorf("failed to validate graph: %w", err)
	}

	err = g.Run(ctx, opt, args)
	if err != nil {
		var errs *dag.Errors
		if errors.As(err, &errs) {
			if len(errs.Errors) == 1 {
				// If we are returning an exit code of 2 when asking for terraform plan's detailed-exitcode then pass that exit code
				var eerr *exec.ExitError
				if detailedExitcode && errors.As(errs.Errors[0], &eerr) && eerr.ExitCode() == 2 {
					Logger.Printf("plan has changes\n")
					return eerr
				}
			}
		}
		return fmt.Errorf("failed to run graph: %w", err)
	}

	return nil
}
