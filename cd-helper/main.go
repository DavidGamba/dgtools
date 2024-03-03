package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	if len(args) < 1 {
		return fmt.Errorf("missing jump argument")
	}

	fsys := os.DirFS("/")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}
	cwd = strings.TrimPrefix(cwd, "/")

	target, err := GetRelDir(fsys, cwd, args[0])
	if err != nil {
		return fmt.Errorf("failed to get rel jump dir: %w", err)
	}
	rel, err := filepath.Rel(target, cwd)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	Logger.Printf("rel: %s", rel)

	fmt.Println("/" + target)

	return nil
}

func GetRelDir(fsys fs.FS, currentDir, jump string) (string, error) {
	target, err := GetJumpDir(fsys, currentDir, jump)
	if err != nil {
		return "", fmt.Errorf("failed to get jump dir: %w", err)
	}

	c := strings.Split(currentDir, "/")
	Logger.Printf("c: %v", c)
	t := strings.Split(target, "/")
	Logger.Printf("t: %v", t)
	f := []string{}
	for i := 0; i < len(c) && i < len(t); i++ {
		if c[i] != t[i] {
			break
		}
		f = append(f, c[i])
	}
	f = append(f, jump)
	if len(c) > len(f) {
		f = append(f, c[len(f):]...)
	}
	Logger.Printf("f: %v", f)

	final := strings.Join(f, "/")
	if fi, err := fs.Stat(fsys, final); err == nil {
		if fi.IsDir() {
			return final, nil
		}
		return "", fmt.Errorf("found %s, but it is not a directory", final)
	}

	return target, nil
}

func GetJumpDir(fsys fs.FS, currentDir, jump string) (string, error) {
	d := filepath.Join(currentDir, "..", jump)
	Logger.Printf("d: %s", d)
	if fi, err := fs.Stat(fsys, d); err == nil {
		if fi.IsDir() {
			return d, nil
		}
		return "", fmt.Errorf("found %s, but it is not a directory", d)
	}
	d = filepath.Join(currentDir, "..", "*"+jump)
	Logger.Printf("d: %s", d)
	if ee, err := fs.Glob(fsys, d); err == nil {
		Logger.Printf("v: %v", ee)
		if len(ee) == 1 {
			if fi, err := fs.Stat(fsys, ee[0]); err == nil {
				if fi.IsDir() {
					return filepath.Join(currentDir, "..", fi.Name()), nil
				}
				return "", fmt.Errorf("found %s, but it is not a directory", d)
			}
		}
	}
	d = filepath.Join(currentDir, "..", jump+"*")
	Logger.Printf("d: %s", d)
	if ee, err := fs.Glob(fsys, d); err == nil {
		Logger.Printf("v: %v", ee)
		if len(ee) == 1 {
			if fi, err := fs.Stat(fsys, ee[0]); err == nil {
				if fi.IsDir() {
					return filepath.Join(currentDir, "..", fi.Name()), nil
				}
				return "", fmt.Errorf("found %s, but it is not a directory", d)
			}
		}
	}
	if currentDir == "." {
		return "", fmt.Errorf("not found")
	}
	d = filepath.Join(currentDir, "..")
	Logger.Printf("d: %s", d)
	return GetJumpDir(fsys, d, jump)
}
