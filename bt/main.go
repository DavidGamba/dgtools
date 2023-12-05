package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/bt/terraform"
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
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	ctx = config.NewConfigContext(ctx, cfg)

	opt := getoptions.New()
	opt.Self("", "Terraform build system built as a no lock-in wrapper")
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.SetUnknownMode(getoptions.Pass)

	terraform.NewCommand(ctx, opt)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}
