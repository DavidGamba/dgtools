package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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

	iexec := opt.NewCommand("iexec", "interactive exec")
	iexec.HelpSynopsisArg("<pod|deployment>", "pod or deployment name")
	iexec.Int("index", 0, opt.Description("index of the container to exec into"))
	iexec.String("namespace", "", opt.Alias("ns"))
	iexec.SetCommandFn(IExec)

	opt.HelpCommand("help", opt.Alias("?"))
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

// Search in pod, deployment, statefulset
func IExec(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	index := opt.Value("index").(int)
	ns := opt.Value("namespace").(string)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <pod|deployment>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	selector := args[0]

	cmdBase := []string{"kubectl"}
	if ns != "" {
		cmdBase = append(cmdBase, []string{"-n", ns}...)
	}
	cmd := append(cmdBase, []string{"get", "pod", selector}...)
	err := run.CMD(cmd...).Log().Run()
	if err == nil {
		return execIntoPod(ctx, selector)
	}

	cmd = append(cmdBase, []string{"get", "deployment", selector, "--output", "jsonpath={.metadata.labels.app}"}...)
	out, err := run.CMD(cmd...).Log().STDOutOutput()
	if err == nil {
		cmd = append(cmdBase, []string{"get", "pod", "-l", fmt.Sprintf("app=%s", string(out)), "-o", "custom-columns=NAME:metadata.name", "--no-headers"}...)
		out, err := run.CMD(cmd...).Log().STDOutOutput()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Found pods:\n%s\n", string(out))
		lines := strings.Split(string(out), "\n")
		return execIntoPod(ctx, lines[index])
	}

	return fmt.Errorf("not found '%s'", selector)
}

func execIntoPod(ctx context.Context, pod string) error {
	cmd := []string{"kubectl", "exec", "--stdin", "--tty", pod, "--", "/bin/bash"}
	return run.CMD(cmd...).Log().Stdin().Run()
}
