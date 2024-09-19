package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/bt/stack"
	stacksConfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	// Read config and store it in context
	cfg, _, err := config.Get(ctx, ".bt.cue")
	if err != nil {
		if !errors.Is(err, buildutils.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
	}
	ctx = config.NewConfigContext(ctx, cfg)

	// Read config and store it in context
	stackCfg, _, err := stacksConfig.Get(ctx, "bt-stacks.cue")
	if err != nil {
		if !errors.Is(err, buildutils.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
	}
	ctx = stacksConfig.NewConfigContext(ctx, stackCfg)

	opt := getoptions.New()
	opt.Self("", "Terraform build system built as a no lock-in wrapper")
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.String("color", "auto", opt.Description("show colored output"), opt.ValidValues("always", "auto", "never"))
	opt.SetUnknownMode(getoptions.Pass)

	configCMD(ctx, opt)
	terraform.NewCommand(ctx, opt)
	stack.NewCommand(ctx, opt)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
		run.Logger.SetOutput(io.Discard)
		config.Logger.SetOutput(io.Discard)
		stack.Logger.SetOutput(io.Discard)
		terraform.Logger.SetOutput(io.Discard)
	}

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if errors.Is(err, getoptions.ErrorParsing) {
			fmt.Fprintf(os.Stderr, "\n%s", opt.Help())
		}
		return 1
	}
	return 0
}
