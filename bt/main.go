package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/bt/stack"
	stacksConfig "github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/cueutils"
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

	os.Setenv("CUE_EXPERIMENT", "embed")

	// Read config and store it in context
	cfgValue := cueutils.NewValue()
	cfg, _, err := config.Get(ctx, cfgValue, ".bt.cue")
	if err != nil {
		if !errors.Is(err, buildutils.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
	}
	ctx = config.NewConfigContext(ctx, cfg)

	// Read config and store it in context
	stackValue := cueutils.NewValue()
	stackCfg, _, err := stacksConfig.Get(ctx, stackValue, "bt-stacks.cue")
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
	opt.Bool("print-raw-config", false)
	opt.String("color", "auto", opt.Description("show colored output"), opt.ValidValues("always", "auto", "never"))
	opt.SetUnknownMode(getoptions.Pass)

	opt.NewCommand("version", "Show version").SetCommandFn(printVersion())

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
		cueutils.Logger.SetOutput(io.Discard)
	}

	if opt.Called("print-raw-config") {
		fmt.Printf("config value:\n%v\n", cfgValue)
		fmt.Printf("stack value:\n%v\n", stackValue)
		return 0
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
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			return eerr.ExitCode()
		}
		var serr *stack.ExitError
		if errors.As(err, &serr) {
			return serr.ExitCode()
		}
		return 1
	}
	return 0
}
