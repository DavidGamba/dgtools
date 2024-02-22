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
	"github.com/DavidGamba/dgtools/buildutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func TestPlan(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestPlan without config", func(t *testing.T) {
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
			if !slices.Equal(r.Cmd, []string{"terraform", "plan", "-out", ".tf.plan", "-no-color"}) {
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
		opt.String("profile", "default")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.String("ws", "")
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := planRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestPlan error: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestPlan with default config but no valid profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(tDir, "environments"), 0755)
		_ = buildutils.Touch(filepath.Join(tDir, "environments", "dev.tfvars"))
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "plan", "-out", ".tf.plan-dev", "-var-file", "/home/user/dev-backend-config.json", "-var-file", "environments/dev.tfvars", "-no-color"}) {
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
		opt.String("profile", "default")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.String("ws", "dev")
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := planRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestPlan error: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestPlan with default config and dev profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(tDir, "environments"), 0755)
		_ = buildutils.Touch(filepath.Join(tDir, "environments", "dev.tfvars"))
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "plan", "-out", ".tf.plan-dev", "-var-file", "/home/user/dev-backend-config.json", "-var-file", "environments/dev.tfvars", "-no-color"}) {
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
		opt.String("profile", "dev")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.String("ws", "dev")
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := planRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestPlan error: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestPlan with default config and prod profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		tDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(tDir, "environments"), 0755)
		_ = buildutils.Touch(filepath.Join(tDir, "environments", "prod.tfvars"))
		ctx = NewDirContext(ctx, tDir)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"terraform", "plan", "-out", ".tf.plan-prod", "-var-file", "/tmp/terraform-project/prod-backend-config.json", "-var-file", "environments/prod.tfvars", "-no-color"}) {
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
		opt.String("profile", "prod")
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("ignore-cache", false)
		opt.String("ws", "prod")
		opt.StringSlice("var-file", 1, 1)
		opt.StringSlice("target", 1, 99)
		opt.StringSlice("replace", 1, 99)
		err := planRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestPlan error: %s", err)
		}
		t.Log(buf.String())
	})
}
