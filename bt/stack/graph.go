package stack

import (
	"context"
	"fmt"
	"os"

	sconfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
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

	cfg := sconfig.ConfigFromContext(ctx)

	wsFn := func(component, dir, ws string, variables []string) getoptions.CommandFn {
		return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
			return nil
		}
	}

	g, err := generateDAG(opt, id, cfg, normal, wsFn)
	if err != nil {
		return err
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
		Logger.Printf("graph saved to: %s\n", filename)
	}

	return nil
}
