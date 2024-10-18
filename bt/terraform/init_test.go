package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/cueutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func TestInit(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestInit without config", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"terraform", "init", "-no-color"}) {
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
		opt.Bool("ignore-cache", false)
		opt.String("profile", "default")
		opt.String("color", "auto")
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.init")); os.IsNotExist(err) {
			t.Errorf("no .tf.init file: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestInit with default config but no valid profile selected", func(t *testing.T) {
		_ = os.Remove(".tf.init")
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "." {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "init", "-backend-config", "/home/user/dev-credentials.json", "-no-color"}) {
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
		opt.Bool("ignore-cache", false)
		opt.String("profile", "default")
		opt.String("color", "auto")
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit error: %s", err)
		}
		if _, err := os.Stat(".tf.init"); os.IsNotExist(err) {
			t.Errorf("no .tf.init file: %s", err)
		}
		_ = os.Remove(".tf.init")
		t.Log(buf.String())
	})

	t.Run("TestInit with default config and dev profile selected", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"tofu", "init", "-backend-config", "/home/user/dev-credentials.json", "-no-color"}) {
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
		opt.Bool("ignore-cache", false)
		opt.String("profile", "dev")
		opt.String("color", "auto")
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.init")); os.IsNotExist(err) {
			t.Errorf("no .tf.init file: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestInit with default config and prod profile selected", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"terraform", "init", "-backend-config", "/tmp/terraform-project/prod-credentials.json", "-no-color"}) {
				return fmt.Errorf("unexpected cmd: %v", r.Cmd)
			}
			for _, e := range r.GetEnv() {
				if strings.Contains(e, "TF_DATA_DIR") {
					if e != "TF_DATA_DIR=.terraform-prod" {
						return fmt.Errorf("unexpected env: %v", e)
					}
				}
			}
			return nil
		})
		ctx = run.ContextWithRunInfo(ctx, mock)
		opt := getoptions.New()
		opt.Bool("dry-run", false)
		opt.Bool("ignore-cache", false)
		opt.String("profile", "prod")
		opt.String("color", "auto")
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.init")); os.IsNotExist(err) {
			t.Errorf("no .tf.init file: %s", err)
		}
		t.Log(buf.String())
	})
}
