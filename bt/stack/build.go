package stack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/DavidGamba/dgtools/bt/config"
	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/go-getoptions"
)

type ExitError struct {
	exitCode int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.exitCode)
}

func (e *ExitError) ExitCode() int {
	return e.exitCode
}

func BuildCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("build", "Builds the stack")
	opt.SetCommandFn(BuildRun)
	opt.Bool("apply", false, opt.Description("Apply Terraform plan"))
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("dry-run", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.Bool("no-checks", false, opt.Description("Do not run pre-apply checks"), opt.Alias("nc"))
	opt.Bool("reverse", false, opt.Description("Reverses the order of operation"))
	opt.Bool("serial", false)
	opt.Bool("show", false, opt.Description("Show Terraform plan"))
	opt.Bool("lock", false, opt.Description("Run 'terraform providers lock' after init"))
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.Int("parallelism", 10*runtime.GOMAXPROCS(0), opt.Description("Pass through to Terraform -parallelism flag"))
	opt.Int("stack-parallelism", runtime.GOMAXPROCS(0), opt.Description("Max number of stack components to run in parallel"))

	return opt
}

func BuildRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	id := opt.Value("id").(string)
	reverse := opt.Value("reverse").(bool)
	serial := opt.Value("serial").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	stackParallelism := opt.Value("stack-parallelism").(int)

	if id == "" {
		fmt.Fprintf(os.Stderr, "ERROR: missing stack id\n")
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	normal := !reverse

	cfg := sconfig.ConfigFromContext(ctx)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	wsFn := func(component, dir, ws string, variables []string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			ctx = terraform.NewComponentContext(ctx, fmt.Sprintf("%s:%s", component, ws))
			ctx = terraform.NewStackContext(ctx, true)
			d := filepath.Join(cfg.ConfigRoot, dir)
			d, err = filepath.Rel(wd, d)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			ctx = terraform.NewDirContext(ctx, d)

			nopt := getoptions.New()
			nopt.Bool("apply", opt.Value("apply").(bool))
			nopt.Bool("destroy", opt.Value("destroy").(bool))
			nopt.Bool("detailed-exitcode", opt.Value("detailed-exitcode").(bool))
			nopt.Bool("dry-run", opt.Value("dry-run").(bool))
			nopt.Bool("ignore-cache", opt.Value("ignore-cache").(bool))
			nopt.Bool("no-checks", opt.Value("no-checks").(bool))
			nopt.Bool("show", opt.Value("show").(bool))
			nopt.Bool("lock", opt.Value("lock").(bool))
			nopt.String("profile", opt.Value("profile").(string))
			nopt.String("color", opt.Value("color").(string))
			nopt.String("ws", ws)
			nopt.StringSlice("replace", 1, 99)
			nopt.StringSlice("target", 1, 99)
			nopt.StringSlice("var", 1, 99)
			nopt.StringSlice("var-file", 1, 1)
			nopt.Int("parallelism", opt.Value("parallelism").(int))
			err = nopt.SetValue("var", variables...)
			if err != nil {
				return fmt.Errorf("failed to set variables: %w", err)
			}

			return terraform.BuildRun(ctx, nopt, args)
		}
	}

	g, err := generateDAG(opt, id, cfg, normal, wsFn)
	if err != nil {
		return err
	}
	g.SetMaxParallel(stackParallelism)
	Logger.Printf("stack parallelism: %d\n", stackParallelism)

	if serial {
		g.SetSerial()
	}

	err = g.Run(ctx, opt, args)
	if err != nil {
		return fmt.Errorf("failed to run graph: %w", err)
	}

	if detailedExitcode && terraform.HasChanges {
		eerr := &ExitError{exitCode: 2}
		return fmt.Errorf("stack has changes: %w", eerr)
	}

	return nil
}
