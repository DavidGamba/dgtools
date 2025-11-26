package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/DavidGamba/dgtools/yamlutils"
	"github.com/DavidGamba/go-getoptions"
	"go.yaml.in/yaml/v4"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("verbose", false)

	opt.SetCommandFn(Run)
	opt.HelpSynopsisArg("<yaml_file>", "yaml file to sort")

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("verbose") {
		Logger.SetOutput(os.Stderr)
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
		if errors.Is(err, getoptions.ErrorParsing) {
			fmt.Fprintf(os.Stderr, "\n%s", opt.Help())
		}
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	filename, _, err := opt.GetRequiredArg(args)
	if err != nil {
		return err
	}

	Logger.Printf("Sorting file %q", filename)
	yaml, err := yamlutils.NewFromFile(filename)
	if err != nil {
		return fmt.Errorf("reading yaml file: %w", err)
	}
	if len(yaml) == 0 {
		return fmt.Errorf("no YAML documents found in file")
	}
	Logger.Printf("yaml: %v", yaml[0].Tree)
	printSortedTree(yaml[0].Tree, 0, false, false)

	return nil
}

func printSortedTree(tree any, level int, arrayKey, arrayElement bool) {
	spacing := level * 3
	if arrayElement {
		spacing = (level * 3) - 2
	}
	switch tree := tree.(type) {
	case map[string]any:
		Logger.Printf("NavigateTree: map[string] type")
		keys := make([]string, len(tree))
		i := 0
		for k := range tree {
			keys[i] = k
			i++
		}
		slices.Sort(keys)
		for _, k := range keys {
			v := tree[k]
			if arrayKey {
				arrayKey = false
				fmt.Printf("%s: ", k)
			} else {
				fmt.Printf("%s%s: ", strings.Repeat(" ", spacing), k)
			}
			switch v.(type) {
			case string, bool, int, int64, float64, nil:
				printSortedTree(v, 0, false, arrayElement)
			default:
				fmt.Println()
				printSortedTree(v, level+1, false, arrayElement)
			}
		}
	case []any:
		Logger.Printf("NavigateTree: slice/array type")
		for _, v := range tree {
			fmt.Printf("%s  - ", strings.Repeat(" ", (level*3)-3))
			printSortedTree(v, level+1, true, true)
		}
	default:
		Logger.Printf("NavigateTree: single element type")
		out, err := yaml.Marshal(tree)
		if err != nil {
			Logger.Printf("Error marshalling: %s", err)
		} else {
			fmt.Print(string(out))
		}
	}
}
