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

	ghmd := opt.NewCommand("github-collapsible-comment", "Create a collapsible comment on GitHub")
	ghmd.SetCommandFn(GHCommentRun)
	// ghmd.Int("part", 1, opt.Description("Part number if the message is bigger than 65536 char limit"))
	ghmd.HelpSynopsisArg("<title>", "Title")

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

// alias pbcopy='xclip -selection clipboard'
// alias pbpaste='xclip -selection clipboard -o'

func GHCommentRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")

	cmd := []string{"pbcopy"}
	input := []byte(`
<details><summary>title</summary>

` + "```" + `
Content
` + "```" + `
</details>
`)
	err := run.CMD(cmd...).Log().In(input).Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	return nil
}
