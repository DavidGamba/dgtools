package stack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
	"github.com/mattn/go-isatty"
)

func generateDAG(opt *getoptions.GetOpt, id string, cfg *sconfig.Config, normal bool) (*dag.Graph, error) {
	color := opt.Value("color").(string)

	tm := dag.NewTaskMap()
	g := dag.NewGraph("stack " + id)

	if color == "always" || (color == "auto" && isatty.IsTerminal(os.Stdout.Fd())) {
		g.UseColor = true
	}

	wd, err := os.Getwd()
	if err != nil {
		return g, fmt.Errorf("failed to get current working directory: %w", err)
	}

	emptyFn := func(dir string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			return nil
		}
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

	for _, c := range cfg.Stack[sconfig.ID(id)].Components {
		cID := string(c.ID)
		variables := []string{}
		for _, v := range c.Variables {
			variables = append(variables, v.String())
		}

		if len(c.Workspaces) > 0 {
			// workspace mode
			tm.Add(cID, emptyFn(cID))
			g.AddTask(tm.Get(cID))
			for _, w := range c.Workspaces {
				wID := fmt.Sprintf("%s:%s", cID, w)
				tm.Add(wID, wsFn(cID, c.Path, w, variables))
				g.AddTask(tm.Get(wID))
				Logger.Printf("adding task %s on %s ws %s vars: %v\n", wID, c.Path, w, variables)

				if normal {
					g.TaskDependsOn(tm.Get(cID), tm.Get(wID))
				} else {
					g.TaskDependsOn(tm.Get(wID), tm.Get(cID))
				}
				if c.Retries > 0 {
					g.TaskRetries(tm.Get(wID), c.Retries)
				}
			}
		} else {
			// normal mode
			tm.Add(cID, wsFn(cID, c.Path, "", variables))
			Logger.Printf("adding task %s on %s vars: %v\n", cID, c.Path, variables)
			g.AddTask(tm.Get(cID))
			if c.Retries > 0 {
				g.TaskRetries(tm.Get(cID), c.Retries)
			}
		}
	}

	for _, c := range cfg.Stack[sconfig.ID(id)].Components {
		cID := string(c.ID)

		if normal {
			for _, e := range c.DependsOn {
				eID := e
				if len(c.Workspaces) > 0 {
					// workspace mode
					for _, w := range c.Workspaces {
						wID := fmt.Sprintf("%s:%s", cID, w)
						g.TaskDependsOn(tm.Get(wID), tm.Get(eID))
					}
				} else {
					// normal mode
					g.TaskDependsOn(tm.Get(cID), tm.Get(eID))
				}
			}
		} else {
			for _, e := range c.DependsOn {
				eID := e
				if len(c.Workspaces) > 0 {
					// workspace mode
					for _, w := range c.Workspaces {
						wID := fmt.Sprintf("%s:%s", cID, w)
						g.TaskDependsOn(tm.Get(eID), tm.Get(wID))
					}
				} else {
					// normal mode
					g.TaskDependsOn(tm.Get(eID), tm.Get(cID))
				}
			}
		}
	}

	err = g.Validate(tm)
	if err != nil {
		return g, fmt.Errorf("failed to build graph: %w", err)
	}

	return g, nil
}
