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

func MirrorCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("mirror", "Creates a mirror of all the providers in the stack")
	opt.SetCommandFn(MirrorRun)
	opt.Bool("dry-run", false)
	opt.Bool("serial", false)
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.Int("parallelism", 10*runtime.GOMAXPROCS(0), opt.Description("Pass through to Terraform -parallelism flag"))
	opt.StringSlice("platform", 1, 99, opt.Description("Target platform"), opt.ArgName("os_arch"))

	return opt
}

func MirrorRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	id := opt.Value("id").(string)
	serial := opt.Value("serial").(bool)

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

			platforms := opt.Value("platform").([]string)

			nopt := getoptions.New()
			nopt.Bool("dry-run", opt.Value("dry-run").(bool))
			nopt.String("profile", opt.Value("profile").(string))
			nopt.String("color", opt.Value("color").(string))
			nopt.String("ws", ws)
			nopt.Int("parallelism", opt.Value("parallelism").(int))
			nopt.StringSlice("platform", 1, 99)

			err = nopt.SetValue("platform", platforms...)
			if err != nil {
				return fmt.Errorf("failed to set platforms: %w", err)
			}

			return terraform.ProvidersMirrorRun(ctx, nopt, args)
		}
	}

	g, err := generateDAG(opt, id, cfg, true, wsFn)
	if err != nil {
		return err
	}
	g.SetSerial()

	if serial {
		g.SetSerial()
	}

	err = g.Run(ctx, opt, args)
	if err != nil {
		return fmt.Errorf("failed to run graph: %w", err)
	}

	return nil
}
