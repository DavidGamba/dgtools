package main

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/DavidGamba/go-getoptions"
)

func printVersion() getoptions.CommandFn {
	return func(context.Context, *getoptions.GetOpt, []string) error {
		fmt.Printf("%15s %s\n", "go.version", runtime.Version())
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("failed to read build info")
		}
		fmt.Printf("%15s %s\n", "module.path", info.Main.Path)
		fmt.Printf("%15s %s\n", "module.version", info.Main.Version)
		for _, s := range info.Settings {
			if s.Value != "" {
				fmt.Printf("%15s %s\n", s.Key, s.Value)
			}
		}
		return nil
	}
}
