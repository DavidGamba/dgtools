package stack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/bt/config"
	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/go-getoptions"
)

func InitCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("init", "Runs terraform init on each component of the stack")
	opt.SetCommandFn(InitRun)
	opt.Bool("dry-run", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.Bool("serial", false)
	opt.Bool("lock", false, opt.Description("Run 'terraform providers lock' after init"))
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.Int("stack-parallelism", 1, opt.Description("Max number of stack components to run in parallel"))
	opt.Bool("tf-in-automation", false, opt.Description(`Determine if we are running in automation.
It will use a separate TF_DATA_DIR per workspace.`), opt.GetEnv("TF_IN_AUTOMATION"), opt.GetEnv("BT_IN_AUTOMATION"))

	return opt
}

func InitRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	id := opt.Value("id").(string)
	serial := opt.Value("serial").(bool)
	stackParallelism := opt.Value("stack-parallelism").(int)

	if id == "" {
		fmt.Fprintf(os.Stderr, "ERROR: missing stack id\n")
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

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
			nopt.Bool("dry-run", opt.Value("dry-run").(bool))
			nopt.Bool("ignore-cache", opt.Value("ignore-cache").(bool))
			nopt.Bool("tf-in-automation", opt.Value("tf-in-automation").(bool))
			nopt.String("profile", opt.Value("profile").(string))
			nopt.String("color", opt.Value("color").(string))
			nopt.String("ws", ws)

			return terraform.InitRun(ctx, nopt, args)
		}
	}

	g, err := generateDAG(opt, id, cfg, true, wsFn)
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

	return nil
}
