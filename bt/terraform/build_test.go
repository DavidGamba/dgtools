package terraform

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/cueutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func TestBuild(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestBuild without config", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		value := cueutils.NewValue()
		cfg, _, _ := config.Get(ctx, value, "x")
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			op := r.Cmd[1]
			switch op {
			case "init":
				if !slices.Equal(r.Cmd, []string{"terraform", "init", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			case "plan":
				if !slices.Equal(r.Cmd, []string{"terraform", "plan", "-out", ".tf.plan", "-parallelism", "10", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			case "apply":
				if !slices.Equal(r.Cmd, []string{"terraform", "apply", "-parallelism", "10", "-input", ".tf.plan-dev", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			default:
				return fmt.Errorf("unexpected cmd: %v", r.Cmd)
			}
			for _, e := range r.GetEnv() {
				if strings.Contains(e, "TF_DATA_DIR") {
					if e != "TF_DATA_DIR=.terraform" {
						return fmt.Errorf("unexpected env: %v", e)
					}
				}
			}
			return nil
		})
		ctx = run.ContextWithRunInfo(ctx, mock)
		opt := getoptions.New()
		opt.Bool("dry-run", false)
		opt.Int("parallelism", 10)
		opt.String("profile", "default")
		opt.String("ws", "")
		opt.String("color", "auto")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.Bool("no-checks", false)
		opt.Bool("apply", false)
		opt.Bool("show", false)
		opt.Bool("lock", false)
		opt.Bool("tf-in-automation", false)
		opt.StringSlice("var", 1, 1)
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := BuildRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestBuild error: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestBuild lock", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			op := r.Cmd[1]
			switch op {
			case "init":
				if !slices.Equal(r.Cmd, []string{"tofu", "init", "-backend-config", "/home/user/dev-credentials.json", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			case "plan":
				if !slices.Equal(r.Cmd, []string{"tofu", "plan", "-out", ".tf.plan-hello", "-parallelism", "10", "-var-file", "/home/user/dev-backend-config.json", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			case "providers":
				if !slices.Equal(r.Cmd, []string{"tofu", "providers", "lock", "-platform=darwin_amd64", "-platform=darwin_arm64", "-platform=linux_amd64", "-platform=linux_arm64", "-no-color"}) {
					return fmt.Errorf("unexpected cmd: %v", r.Cmd)
				}
			default:
				return fmt.Errorf("unexpected cmd: %v", r.Cmd)
			}
			for _, e := range r.GetEnv() {
				if strings.Contains(e, "TF_DATA_DIR") {
					if e != "TF_DATA_DIR=.terraform" {
						return fmt.Errorf("unexpected env: %v", e)
					}
				}
			}
			return nil
		})
		ctx = run.ContextWithRunInfo(ctx, mock)
		opt := getoptions.New()
		opt.Bool("dry-run", false)
		opt.Int("parallelism", 10)
		opt.String("profile", "default")
		opt.String("ws", "hello")
		opt.String("color", "auto")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.Bool("no-checks", false)
		opt.Bool("apply", false)
		opt.Bool("show", false)
		opt.Bool("lock", true)
		opt.Bool("tf-in-automation", false)
		opt.StringSlice("var", 1, 1)
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := BuildRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestBuild error: %s", err)
		}
		t.Log(buf.String())
	})
}
