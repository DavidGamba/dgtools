package run_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/DavidGamba/dgtools/run"
)

func TestRunWithMocks(t *testing.T) {
	t.Run("mock", func(t *testing.T) {
		r := run.CMD("ls", "./run")
		r.Mock(func(r *run.RunInfo) error {
			r.Stdout.Write([]byte("hello world\n"))
			r.Stderr.Write([]byte("hola mundo\n"))
			return nil
		})
		out, err := r.CombinedOutput()
		if err != nil {
			t.Errorf("unexpected error")
		}
		if string(out) != "hello world\nhola mundo\n" {
			t.Errorf("wrong output: %s\n", out)
		}
	})

	t.Run("mock with context", func(t *testing.T) {
		ctx := context.Background()
		mockR := run.CMD().Mock(func(r *run.RunInfo) error {
			r.Stdout.Write([]byte("hello world\n"))
			r.Stderr.Write([]byte("hola mundo\n"))
			return nil
		})
		ctx = run.ContextWithRunInfo(ctx, mockR)

		r := run.CMDCtx(ctx, "ls", "./run")
		out, err := r.CombinedOutput()
		if err != nil {
			t.Errorf("unexpected error")
		}
		if string(out) != "hello world\nhola mundo\n" {
			t.Errorf("wrong output: %s\n", out)
		}
	})

	t.Run("mock with context switch", func(t *testing.T) {
		ctx := context.Background()
		mockR := run.CMD().Mock(func(r *run.RunInfo) error {
			cmd := r.Cmd

			switch {
			case slices.Compare(cmd, []string{"ls", "./run"}) == 0:
				r.Stdout.Write([]byte("hello world\n"))
				r.Stderr.Write([]byte("hola mundo\n"))
				return nil
			case slices.Compare(cmd, []string{"ls", "x"}) == 0:
				r.Stderr.Write([]byte("not found x\n"))
				return fmt.Errorf("not found x")
			default:
				return fmt.Errorf("unexpected command: %s", cmd)
			}
		})
		ctx = run.ContextWithRunInfo(ctx, mockR)

		r := run.CMDCtx(ctx, "ls", "./run")
		out, err := r.CombinedOutput()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if string(out) != "hello world\nhola mundo\n" {
			t.Errorf("wrong output: %s\n", out)
		}

		r = run.CMDCtx(ctx, "ls", "x")
		out, err = r.CombinedOutput()
		if err == nil {
			t.Errorf("expected error")
		}
		if string(out) != "not found x\n" {
			t.Errorf("wrong output: %s\n", out)
		}
	})
}
