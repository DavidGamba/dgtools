package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.HelpCommand("help", opt.Alias("?"))

	it := opt.NewCommand("it", "Run interactively")
	it.HelpSynopsisArg("image", "docker image")
	it.SetCommandFn(Run)

	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

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

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing image\n")
		return getoptions.ErrorHelpCalled
	}
	image := args[0]
	args = args[1:]

	// Docker command to run interactively
	cmd := []string{"docker", "run"}
	cmd = append(cmd, args...)
	cmd = append(cmd, "--entrypoint", "/bin/bash", "-it", "--rm", image)
	err := run.CMD(cmd...).Log().Stdin().Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	return nil
}
