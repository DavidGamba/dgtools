package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/trees"
	"github.com/DavidGamba/dgtools/yamlutils"
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
	opt.SetCommandFn(Run)
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if errors.Is(err, getoptions.ErrorParsing) {
			fmt.Fprintf(os.Stderr, "\n"+opt.Help())
		}
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	// TODO: benchmark tree walking vs struct unmarshalling for the fun of it
	ymlList, err := yamlutils.NewFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get kube config: %w", err)
	}
	if len(ymlList) < 1 {
		return fmt.Errorf("no resources found in kubeconfig")
	}
	current, err := ymlList[0].GetString(false, []string{"current-context"})
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	kctxsI, _, err := trees.NavigateTree(false, ymlList[0].Tree, []string{"contexts"})
	if err != nil {
		return fmt.Errorf("failed to get contexts: %w", err)
	}
	kctxs, ok := kctxsI.([]interface{})
	if !ok {
		return fmt.Errorf("failed to convert contexts: %w", err)
	}

	for _, kctx := range kctxs {
		name := kctx.(map[interface{}]interface{})["name"].(string)
		out := name
		if name != current {
			continue
		}
		c := kctx.(map[interface{}]interface{})["context"]
		namespace := c.(map[interface{}]interface{})["namespace"]
		if namespace != nil {
			out += "/" + namespace.(string)
		}
		fmt.Println(out)
	}

	return nil
}
