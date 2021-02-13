// This file is part of watchmake.
//
// Copyright (C) 2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package main provides a utility to watch the filesystem for changes and run tasks when that happens.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/dgtools/private/hclutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/dgtools/watchmake/lib"
	"github.com/DavidGamba/go-getoptions"
	"github.com/fsnotify/fsnotify"
)

// Logger - default logger instance
var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

var configFileName = "watchmake.hcl"

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false, opt.GetEnv("DEBUG"))
	opt.Bool("validate", false)
	opt.SetUnknownMode(getoptions.Pass)
	opt.NewCommand("cmd", "description").SetCommandFn(Run)
	opt.HelpCommand("")
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	ctx, cancel, done := opt.InterruptContext()
	defer func() { cancel(); <-done }()

	if opt.Called("validate") {
		// config, err := ReadConfig()
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		// 	return 1
		// }
		// fmt.Println(config)
		return 0
	}

	if len(remaining) == 0 {
		err := all()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			return 1
		}
		return 0
	}

	err = opt.Dispatch(ctx, "help", remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

// Run -
func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Running")
	return nil
}

func all() error {
	parser, f, err := hclutils.ParseHCLFile(os.Stderr, configFileName)
	if err != nil {
		return err
	}
	config, err := lib.DecodeConfig(os.Stderr, parser, f)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for _, w := range config.Watchers {
		for _, file := range w.Output {
			err = watcher.Add(file)
			if err != nil {
				return err
			}
		}
	}

	done := make(chan bool)
	go func(config *lib.Watchmake) {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Chmod == fsnotify.Chmod {
					log.Println("modified file:", event.Name)
					for _, t := range config.Tasks {
						for _, taskExecution := range t.Tasks {
							for _, cmd := range taskExecution.Actions {
								for i, e := range cmd {
									if e == "{}" {
										cmd[i] = event.Name
									}
								}
								fmt.Println("$ " + strings.Join(cmd, " "))
								out, err := run.CMD(cmd...).CombinedOutput()
								fmt.Println("> " + string(out))
								if err != nil {
									log.Println("ERROR: ", err)
								}
							}
						}
					}
				}
				// Automatically re-add deleted files
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					err = watcher.Add(event.Name)
					if err != nil {
						log.Println("ERROR: ", err)
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}(config)

	<-done
	return nil
}
