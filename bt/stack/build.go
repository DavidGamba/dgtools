package stack

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/DavidGamba/dgtools/bt/config"
	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/go-getoptions"
)

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
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.Int("parallelism", 10*runtime.NumCPU())

	return opt
}

func BuildRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	id := opt.Value("id").(string)
	reverse := opt.Value("reverse").(bool)
	serial := opt.Value("serial").(bool)

	if id == "" {
		fmt.Fprintf(os.Stderr, "ERROR: missing stack id\n")
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	normal := !reverse

	cfg := sconfig.ConfigFromContext(ctx)

	g, err := generateDAG(opt, id, cfg, normal)
	if err != nil {
		return err
	}
	g.SetMaxParallel(runtime.NumCPU())

	if serial {
		g.SetSerial()
	}

	err = g.Run(ctx, opt, args)
	if err != nil {
		return fmt.Errorf("failed to run graph: %w", err)
	}

	return nil
}
