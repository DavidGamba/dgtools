package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	opt.HelpSynopsisArgs("<file>...")
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.String("namespace", "")
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
	Logger.Printf("Running")
	namespace := opt.Value("namespace").(string)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	file := args[0]
	prefix := strings.TrimSuffix(strings.TrimSuffix(file, ".yaml"), ".yml")

	cmd := []string{"yaml-seam", "split", file, "-d", prefix, "--force"}
	err := run.CMD(cmd...).Log().Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	entries, err := fs.ReadDir(os.DirFS(prefix), ".")
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fn := filepath.Join(prefix, e.Name())
		if !strings.HasSuffix(fn, ".yaml") {
			continue
		}
		pn := strings.TrimSuffix(fn, ".yaml") + ".patch"
		dn := strings.TrimSuffix(fn, ".yaml") + ".dump"

		Logger.Printf("file: %s\n", fn)
		kname, err := run.CMD("yaml-parse", fn, "-k", "metadata/name", "-q").Log().STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		kkind, err := run.CMD("yaml-parse", fn, "-k", "kind", "-q").Log().STDOutOutput()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		Logger.Printf("%s resource: %s %s\n", fn, kkind, kname)

		dh, err := os.Create(dn)
		if err != nil {
			return fmt.Errorf("failed to create patch file: %w", err)
		}
		defer dh.Close()

		cmd := []string{"kubectl", "get", "-o", "yaml", string(kkind), string(kname)}
		if namespace != "" {
			cmd = append(cmd, "-n", namespace)
		}

		// TODO: Use k8s apis directly
		serr := ""
		buf := bytes.NewBufferString(serr)
		err = run.CMD(cmd...).Log().Run(dh, buf)
		if err != nil {
			Logger.Printf("Output: %s\n", buf.String())
			// Don't fail if resource needs to be created
			if !strings.Contains(buf.String(), "Error from server (NotFound)") {
				Logger.Printf("ERROR x: %s\n", err.Error())
				return fmt.Errorf("failed to get dump for %s: %w", fn, err)
			}
		}

		retries := 3
		for i := 0; i < retries; i++ {
			time.Sleep(time.Duration(i) * time.Second)
			ph, err := os.Create(pn)
			if err != nil {
				return fmt.Errorf("failed to create patch file: %w", err)
			}
			defer ph.Close()
			cmd = []string{"kubectl", "diff", "-f", fn}
			if namespace != "" {
				cmd = append(cmd, "-n", namespace)
			}
			err = run.CMD(cmd...).Log().Run(ph, os.Stderr)
			if err == nil {
				break
			}
			if err != nil {
				var eerr *exec.ExitError
				if errors.As(err, &eerr) && eerr.ExitCode() == 1 {
					Logger.Printf("%s diff changes found\n", fn)
					break
				}
				if i+1 < retries {
					Logger.Printf("RETRYING: failed to get diff for %s: %s", fn, err)
					continue
				}
				return fmt.Errorf("failed to get diff for %s: %w", fn, err)
			}
		}
	}

	dd, err := filepath.Glob(prefix + "/*.dump")
	if err != nil {
		return fmt.Errorf("failed to expand dump files: %w", err)
	}
	cmd = []string{"yaml-seam", "join", "-o", prefix + "-current.yaml"}
	cmd = append(cmd, dd...)
	err = run.CMD(cmd...).Log().Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	entries, err = fs.ReadDir(os.DirFS(prefix), ".")
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		pn := filepath.Join(prefix, e.Name())
		if !strings.HasSuffix(pn, ".patch") {
			continue
		}
		err := run.CMD("patch-seam", "select", pn).Log().Stdin().Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	entries, err = fs.ReadDir(os.DirFS(prefix), ".")
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		pn := filepath.Join(prefix, e.Name())
		if !strings.HasSuffix(pn, ".fpatch") {
			continue
		}
		dn := strings.TrimSuffix(pn, ".fpatch") + ".dump"

		pc, err := os.ReadFile(pn)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", dn, err)
		}

		err = run.CMD("patch", dn).Log().In(pc).Run()
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
	}

	cmd = []string{"yaml-seam", "join", "-o", prefix + "-patched.yaml"}
	cmd = append(cmd, dd...)
	err = run.CMD(cmd...).Log().Run()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	return nil
}
