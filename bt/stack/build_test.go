package stack

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/DavidGamba/dgtools/bt/stack/config"
	"github.com/DavidGamba/dgtools/cueutils"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	Logger.SetOutput(buf)
	run.Logger.SetOutput(buf)
	cueutils.Logger.SetOutput(buf)
	return buf
}

func TestStack(t *testing.T) {
	t.Setenv("HOME", "/home/user")

	t.Run("TestStack no id", func(t *testing.T) {
		buf := setupLogging()
		ctx := context.Background()
		opt := getoptions.New()
		opt.String("id", "")
		opt.String("color", "auto")
		opt.Bool("apply", false)
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("dry-run", false)
		opt.Bool("ignore-cache", false)
		opt.Bool("no-checks", false)
		opt.Bool("reverse", false)
		opt.Bool("serial", false)
		opt.Bool("show", false)
		opt.Bool("lock", false)
		opt.String("profile", "default")
		opt.Int("parallelism", 10)
		opt.Int("stack-parallelism", 4)

		err := BuildRun(ctx, opt, []string{})
		if err == nil {
			t.Errorf("Error was expected")
		}
		t.Log(buf.String())
	})

	t.Run("TestStack without config", func(t *testing.T) {

		c := `
package bt_stacks

component: hola: {}
component: hello: {}

stack: x: {
	components: [component.hola, component.hello]
}
`

		r := strings.NewReader(c)

		buf := setupLogging()
		ctx := context.Background()
		value := cueutils.NewValue()
		cfg, err := config.Read(ctx, value, "x.cue", r)
		if err != nil {
			t.Fatalf("failed to read config: %s", err)
		}
		ctx = config.NewConfigContext(ctx, cfg)
		Logger.Printf("config: %v", value)
		tDir := t.TempDir()
		mock := run.CMDCtx(ctx).Mock(func(r *run.RunInfo) error {
			if r.GetDir() != tDir {
				return fmt.Errorf("unexpected dir: %s", r.GetDir())
			}
			if !slices.Equal(r.Cmd, []string{"terraform", "apply", "-parallelism", "10", "-input", ".tf.plan-dev", "-no-color"}) {
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
		opt.String("id", "hello")
		opt.String("color", "auto")
		opt.Bool("apply", false)
		opt.Bool("destroy", false)
		opt.Bool("detailed-exitcode", false)
		opt.Bool("dry-run", false)
		opt.Bool("ignore-cache", false)
		opt.Bool("no-checks", false)
		opt.Bool("reverse", false)
		opt.Bool("serial", false)
		opt.Bool("show", false)
		opt.Bool("lock", false)
		opt.String("profile", "default")
		opt.Int("parallelism", 10)
		opt.Int("stack-parallelism", 4)

		err = BuildRun(ctx, opt, []string{})
		if err != nil {
			t.Errorf("TestStack error: %s", err)
		}
		t.Log(buf.String())
	})
}
