package stack

import (
	"context"
	"fmt"

	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

func generateDAG(id string, cfg *sconfig.Config, normal bool) (*dag.Graph, error) {
	tm := dag.NewTaskMap()
	g := dag.NewGraph("stack " + id)

	emptyFn := func(dir string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			return nil
		}
	}
	normalFn := func(component, dir string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			ctx = terraform.NewComponentContext(ctx, component)
			ctx = terraform.NewDirContext(ctx, dir)
			return terraform.BuildRun(ctx, opt, args)
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

		if len(c.Workspaces) > 0 {
			// workspace mode
			tm.Add(cID, emptyFn(cID))
			g.AddTask(tm.Get(cID))
			for _, w := range c.Workspaces {
				wID := fmt.Sprintf("%s:%s", cID, w)
				tm.Add(wID, wsFn(cID, c.Path, w))
				g.AddTask(tm.Get(wID))

				if normal {
					g.TaskDependensOn(tm.Get(cID), tm.Get(wID))
				} else {
					g.TaskDependensOn(tm.Get(wID), tm.Get(cID))
				}
			}
		} else {
			// normal mode
			tm.Add(cID, normalFn(cID, c.Path))
			g.AddTask(tm.Get(cID))
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
						g.TaskDependensOn(tm.Get(wID), tm.Get(eID))
					}
				} else {
					// normal mode
					g.TaskDependensOn(tm.Get(cID), tm.Get(eID))
				}
			}
		} else {
			for _, e := range c.DependsOn {
				eID := e
				if len(c.Workspaces) > 0 {
					// workspace mode
					for _, w := range c.Workspaces {
						wID := fmt.Sprintf("%s:%s", cID, w)
						g.TaskDependensOn(tm.Get(eID), tm.Get(wID))
					}
				} else {
					// normal mode
					g.TaskDependensOn(tm.Get(eID), tm.Get(cID))
				}
			}
		}
	}

	err := g.Validate(tm)
	if err != nil {
		return g, fmt.Errorf("failed to build graph: %w", err)
	}

	return g, nil
}
