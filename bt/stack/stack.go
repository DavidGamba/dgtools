package stack

import (
	"context"
	"log"
	"os"

	"github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func NewCommand(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("stack", "stack related tasks")

	stacks := []string{}
	for k := range cfg.Stack {
		stacks = append(stacks, string(k))
	}
	opt.String("id", "", opt.ValidValues(stacks...), opt.Description("Stack ID"))

	ConfigCMD(ctx, opt)
	GraphCMD(ctx, opt)
	return opt
}
