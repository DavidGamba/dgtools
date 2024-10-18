package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
	"github.com/mattn/go-isatty"
)

var HasChanges = false

func buildCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("build", "Wraps init, plan and apply into a single operation with a cache")
	opt.SetCommandFn(BuildRun)
	opt.Bool("apply", false, opt.Description("Apply Terraform plan"))
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("dry-run", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the init and plan"), opt.Alias("ic"))
	opt.Bool("no-checks", false, opt.Description("Do not run pre-apply/post-apply checks"), opt.Alias("nc"))
	opt.Bool("show", false, opt.Description("Show Terraform plan"))
	opt.Bool("lock", false, opt.Description("Run 'terraform providers lock' after init"))
	opt.Int("parallelism", 10*runtime.NumCPU())
	opt.StringSlice("replace", 1, 99)
	opt.StringSlice("target", 1, 99)
	opt.StringSlice("var", 1, 99)
	opt.StringSlice("var-file", 1, 1)

	return opt
}

func BuildRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	apply := opt.Value("apply").(bool)
	show := opt.Value("show").(bool)
	lock := opt.Value("lock").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	ws := opt.Value("ws").(string)
	color := opt.Value("color").(string)

	cfg := config.ConfigFromContext(ctx)
	component := ComponentFromContext(ctx)
	dir := DirFromContext(ctx)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	if cfg.TFProfile[cfg.Profile(profile)].Workspaces.Enabled {
		if !workspaceSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)) {
			if ws == "" {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
		}
	}

	lockFn := func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		lockFile := filepath.Join(dir, ".tf.lock")
		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			nopt := getoptions.New()
			nopt.String("profile", opt.Value("profile").(string))
			nopt.StringSlice("platform", 1, 99)
			nopt.String("ws", ws)
			return providersLockRun(ctx, nopt, args)
		}
		return nil
	}

	tm := dag.NewTaskMap()
	tm.Add("init", initRun)
	if lock {
		tm.Add("lock", lockFn)
	}
	tm.Add("plan", planRun)
	if cfg.TFProfile[cfg.Profile(profile)].PreApplyChecks.Enabled {
		tm.Add("checks", checksRun)
	}
	if show {
		tm.Add("show", showPlanRun)
	}
	if apply {
		tm.Add("apply", applyRun)
	}
	if cfg.TFProfile[cfg.Profile(profile)].PostApplyChecks.Enabled {
		tm.Add("post-checks", postChecksRun)
	}

	g := dag.NewGraph(fmt.Sprintf("%s:build", component))

	if color == "always" || (color == "auto" && isatty.IsTerminal(os.Stdout.Fd())) {
		g.UseColor = true
	}

	g.TaskDependsOn(tm.Get("plan"), tm.Get("init"))
	if lock {
		g.TaskDependsOn(tm.Get("lock"), tm.Get("init"))
		g.TaskDependsOn(tm.Get("plan"), tm.Get("lock"))
	}
	if cfg.TFProfile[cfg.Profile(profile)].PreApplyChecks.Enabled {
		g.TaskDependsOn(tm.Get("checks"), tm.Get("plan"))
	}

	if show {
		g.TaskDependsOn(tm.Get("show"), tm.Get("plan"))
	}
	if apply {
		g.TaskDependsOn(tm.Get("apply"), tm.Get("plan"))
		if cfg.TFProfile[cfg.Profile(profile)].PreApplyChecks.Enabled {
			g.TaskDependsOn(tm.Get("apply"), tm.Get("checks"))
		}
		if cfg.TFProfile[cfg.Profile(profile)].PostApplyChecks.Enabled {
			g.TaskDependsOn(tm.Get("post-checks"), tm.Get("apply"))
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
