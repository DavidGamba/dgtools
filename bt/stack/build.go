package stack

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

func BuildCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("build", "Builds the stack")
	opt.SetCommandFn(BuildRun)
	opt.Bool("dry-run", false)
	opt.Bool("serial", false)
	opt.Bool("reverse", false, opt.Description("Reverses the order of operation"))
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("ignore-cache", false, opt.Description("Ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.Bool("no-checks", false, opt.Description("Do not run pre-apply checks"), opt.Alias("nc"))
	opt.Bool("apply", false, opt.Description("Apply Terraform plan"))
	opt.Bool("show", false, opt.Description("Show Terraform plan"))
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.StringSlice("var-file", 1, 1)
	opt.StringSlice("target", 1, 99)
	opt.StringSlice("replace", 1, 99)
	opt.String("ws", "", opt.Description("Workspace to use"))

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

	tm := dag.NewTaskMap()
	g := dag.NewGraph("stack " + id)

	fn := func(dir string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			return nil
		}
	}
	wsFn := func(component, dir, ws string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			ctx = terraform.NewComponentContext(ctx, fmt.Sprintf("%s:%s", component, ws))
			ctx = terraform.NewDirContext(ctx, dir)
			err := opt.SetValue("ws", ws)
			if err != nil {
				return fmt.Errorf("failed to set workspace: %w", err)
			}
			return terraform.BuildRun(ctx, opt, args)
		}
	}

	for _, c := range cfg.Stack[sconfig.ID(id)].Components {
		cID := string(c.ID)
		tm.Add(cID, fn(cID))
		g.AddTask(tm.Get(cID))
		for _, w := range c.Workspaces {
			wID := fmt.Sprintf("%s:%s", c.ID, w)
			tm.Add(wID, wsFn(string(c.ID), c.Path, w))
			g.AddTask(tm.Get(wID))

			if normal {
				g.TaskDependensOn(tm.Get(cID), tm.Get(wID))
			} else {
				g.TaskDependensOn(tm.Get(wID), tm.Get(cID))
			}
		}
	}

	for _, c := range cfg.Stack[sconfig.ID(id)].Components {
		if normal {
			for _, e := range c.DependsOn {
				eID := e
				for _, w := range c.Workspaces {
					wID := fmt.Sprintf("%s:%s", c.ID, w)
					g.TaskDependensOn(tm.Get(wID), tm.Get(eID))
				}
			}
		} else {
			for _, e := range c.DependsOn {
				eID := e
				for _, w := range c.Workspaces {
					wID := fmt.Sprintf("%s:%s", c.ID, w)
					g.TaskDependensOn(tm.Get(eID), tm.Get(wID))
				}
			}
		}
	}

	err := g.Validate(tm)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	if serial {
		g.SetSerial()
	}

	err = g.Run(ctx, opt, args)
	if err != nil {
		return fmt.Errorf("failed to run graph: %w", err)
	}

	return nil
}
