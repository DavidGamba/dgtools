package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	mermaid_go "github.com/dreampuf/mermaid.go"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Render mermaid diagram files to SVG or PNG on the CLI

# Read Mermaid file
mermaid render <filename.mermaid> -o <filename.[svg|png]>

# Pipe Mermaid file to mermaid
cat <filename.mermaid> | mermaid render -o <filename.[svg|png]>
`)

	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.SetUnknownMode(getoptions.Pass)

	render := opt.NewCommand("render", "Render a mermaid diagram to svg or png").SetCommandFn(Render)
	render.String("output", "", opt.Description("Output file.\nUse .svg or .png extension to determine output format."), opt.Required(), opt.ArgName("filename.[svg|png]"))
	render.HelpSynopsisArg("[<filename.mermaid>]", "mermaid input file")

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
		return 1
	}
	return 0
}

func Render(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	outputFile := opt.Value("output").(string)

	re, err := mermaid_go.NewRenderEngine(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize render engine: %w", err)
	}
	defer re.Cancel()

	var reader io.Reader
	if len(args) < 1 {
		// Check if stdin is pipe p or device D
		statStdin, _ := os.Stdin.Stat()
		stdinIsDevice := (statStdin.Mode() & os.ModeDevice) != 0

		if stdinIsDevice {
			fmt.Fprint(os.Stderr, opt.Help())
			return getoptions.ErrorHelpCalled
		}
		Logger.Printf("Reading from stdin\n")
		reader = os.Stdin
	} else {
		filename := args[0]
		Logger.Printf("Reading from file %s\n", filename)
		fh, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fh.Close()
		reader = fh
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read contents: %w", err)
	}

	var rendered []byte
	switch {
	case strings.HasSuffix(outputFile, ".svg"):
		// TODO: Refactor Render to return []byte
		renderedString, err := re.Render(string(bytes))
		if err != nil {
			return fmt.Errorf("failed to render svg: %w", err)
		}
		rendered = []byte(renderedString)
	case strings.HasSuffix(outputFile, ".png"):
		// TODO: Refactor RenderAsPng to return []byte
		rendered, _, err = re.RenderAsPng(string(bytes))
		if err != nil {
			return fmt.Errorf("failed to render png: %w", err)
		}
	default:
		return fmt.Errorf("unknown output file extension, use svg or png: %s", outputFile)
	}

	err = os.WriteFile(outputFile, []byte(rendered), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
