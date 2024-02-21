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
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func TestApply(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestApply without config", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg, _, _ := config.Get(ctx, "x")
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"terraform", "apply", "-input", ".tf.plan-dev", "-no-color"}) {
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
		opt.String("profile", "default")
		opt.String("ws", "dev")
		err := applyRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestApply error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.apply-dev")); os.IsNotExist(err) {
			t.Errorf("no .tf.apply file: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestApply with default config but no valid profile selected", func(t *testing.T) {
		_ = os.Remove(".tf.apply")
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "." {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "apply", "-input", ".tf.plan-dev", "-no-color"}) {
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
		opt.String("profile", "default")
		opt.String("ws", "dev")
		err := applyRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestApply error: %s", err)
		}
		if _, err := os.Stat(".tf.apply-dev"); os.IsNotExist(err) {
			t.Errorf("no .tf.apply file: %s", err)
		}
		_ = os.Remove(".tf.apply")
		t.Log(buf.String())
	})

	t.Run("TestApply with default config and dev profile selected", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"tofu", "apply", "-input", ".tf.plan-dev", "-no-color"}) {
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
		opt.String("profile", "dev")
		opt.String("ws", "dev")
		err := applyRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestApply error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.apply-dev")); os.IsNotExist(err) {
			t.Errorf("no .tf.apply file: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestApply with default config and prod profile selected", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"terraform", "apply", "-input", ".tf.plan-prod", "-no-color"}) {
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
		opt.String("profile", "prod")
		opt.String("ws", "prod")
		err := applyRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestApply error: %s", err)
		}
		if _, err := os.Stat(filepath.Join(tDir, ".tf.apply-prod")); os.IsNotExist(err) {
			t.Errorf("no .tf.apply file: %s", err)
		}
		t.Log(buf.String())
	})
}
