package terraform

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func TestInit(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestInit without config", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg, _, _ := config.Get(ctx, "x")
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "" {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"terraform", "init", "-no-color"}) {
				return fmt.Errorf("unexpected config: %v", r.Cmd)
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
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestInit with default config but no valid profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "" {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "init", "-backend-config", "/home/user/dev-credentials.json", "-no-color"}) {
				return fmt.Errorf("unexpected config: %v", r.Cmd)
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
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestInit with default config and dev profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "" {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"tofu", "init", "-backend-config", "/home/user/dev-credentials.json", "-no-color"}) {
				return fmt.Errorf("unexpected config: %v", r.Cmd)
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
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit: %s", err)
		}
		t.Log(buf.String())
	})

	t.Run("TestInit with default config and prod profile selected", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		cfg := getDefaultConfig()
		ctx = config.NewConfigContext(ctx, cfg)
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != "" {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"terraform", "init", "-backend-config", "/tmp/terraform-project/prod-credentials.json", "-no-color"}) {
				return fmt.Errorf("unexpected config: %v", r.Cmd)
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
		err := initRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestInit: %s", err)
		}
		t.Log(buf.String())
	})
}
