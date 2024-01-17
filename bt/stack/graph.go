package stack

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-getoptions/dag"
)

func GraphCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("graph", "Visual representation of the project layers DAG")
	opt.SetCommandFn(GraphRun)
	opt.Bool("reverse", false, opt.Description("Reverses the order of operation"))
	opt.String("T", "png", opt.Description("Set output format. For example: -T png"))
	opt.String("filename", "", opt.Description("Set output filename"))

	return opt
}

func GraphRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	id := opt.Value("id").(string)
	reverse := opt.Value("reverse").(bool)
	format := opt.Value("T").(string)
	filename := opt.Value("filename").(string)

	if id == "" {
		fmt.Fprintf(os.Stderr, "ERROR: missing stack id\n")
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	if filename == "" {
		filename = fmt.Sprintf("stack-%s.%s", id, format)
	}

	normal := !reverse

	cfg := config.ConfigFromContext(ctx)

	tm := dag.NewTaskMap()
	g := dag.NewGraph("stack " + id)

	fn := func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		return nil
	}

	for _, c := range cfg.Stack[config.ID(id)].Components {
		cID := fmt.Sprintf("c: %s", c.ID)
		tm.Add(cID, fn)
		g.AddTask(tm.Get(cID))
		for _, w := range c.Workspaces {
			wID := fmt.Sprintf("w: %s-%s", c.ID, w)
			tm.Add(wID, fn)
			g.AddTask(tm.Get(wID))

			if normal {
				g.TaskDependensOn(tm.Get(cID), tm.Get(wID))
			} else {
				g.TaskDependensOn(tm.Get(wID), tm.Get(cID))
			}
		}
	}

	for _, c := range cfg.Stack[config.ID(id)].Components {
		cID := fmt.Sprintf("c: %s", c.ID)
		if normal {
			for _, e := range c.DependsOn {
				eID := fmt.Sprintf("c: %s", e)
				g.TaskDependensOn(tm.Get(cID), tm.Get(eID))
			}
		} else {
			for _, e := range c.DependsOn {
				eID := fmt.Sprintf("c: %s", e)
				g.TaskDependensOn(tm.Get(eID), tm.Get(cID))
			}
		}
	}

	err := g.Validate(tm)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Printf("%s\n", g)
	if opt.Called("T") {
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		err = run.CMDCtx(ctx, "dot", "-T", format).Log().In([]byte(g.String())).Run(f, os.Stderr)
		if err != nil {
			return fmt.Errorf("failed to generate graph: %w", err)
		}
	}

	return nil
}
