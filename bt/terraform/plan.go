package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/mattn/go-isatty"
)

func planCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("plan", "")
	opt.Bool("dry-run", false)
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("ignore-cache", false, opt.Description("ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.StringSlice("var-file", 1, 1)
	opt.StringSlice("var", 1, 99)
	opt.StringSlice("target", 1, 99)
	opt.StringSlice("replace", 1, 99)
	opt.SetCommandFn(planRun)

	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	destroy := opt.Value("destroy").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	ignoreCache := opt.Value("ignore-cache").(bool)
	varFiles := opt.Value("var-file").([]string)
	variables := opt.Value("var").([]string)
	targets := opt.Value("target").([]string)
	replacements := opt.Value("replace").([]string)
	ws := opt.Value("ws").(string)

	cfg := config.ConfigFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)
	os.Setenv("CONFIG_ROOT", cfg.ConfigRoot)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	ws, err = getWorkspace(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	defaultVarFiles, err := getDefaultVarFiles(cfg, profile)
	if err != nil {
		return err
	}

	varFiles, err = AddVarFileIfWorkspaceSelected(cfg, profile, dir, ws, varFiles)
	if err != nil {
		return err
	}

	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}

	cwd, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}

	moduleFiles := []string{}
	moduleInfo, diags := tfconfig.LoadModule(".")
	if diags.HasErrors() {
		return fmt.Errorf("failed to load module: %w", diags)
	}
	for module, moduleCall := range moduleInfo.ModuleCalls {
		source := moduleCall.Source
		if _, err := os.Stat(filepath.Join(".", source)); os.IsNotExist(err) {
			Logger.Printf("remote module: %s, %s\n", module, source)
			continue
		}
		// Logger.Printf("local module: %s, %s\n", module, source)
		// include all files in module dir, these could be included templates or scripts.
		moduleFiles = append(moduleFiles, filepath.Join(cwd, source, "*"))
	}

	sources := append(append(append([]string{".tf.init"}, defaultVarFiles...), varFiles...), moduleFiles...)
	sources = append(sources, "./*") // include all files in current dir, these could be included templates or scripts.
	relSources := []string{}
	for _, s := range sources {
		if strings.HasPrefix(s, "/") {
			relSources = append(relSources, filepath.Join("./", s))
		} else {
			relSources = append(relSources, filepath.Join("./", cwd, s))
		}
	}

	filteredSources := []string{}
	globs, _, err := fsmodtime.Glob(os.DirFS("/"), false, relSources)
	if err != nil {
		return fmt.Errorf("failed to glob sources: %w", err)
	}

	for _, g := range globs {
		// Logger.Printf("glob: %s\n", g)
		if !strings.Contains(g, "/.tf.plan") &&
			!strings.Contains(g, "/.tf.check") &&
			!strings.Contains(g, "/.tf.apply") &&
			!strings.Contains(g, "/.terraform/") &&
			!strings.Contains(g, "/.terraform.lock.hcl") {
			filteredSources = append(filteredSources, g)
		}
	}

	// Logger.Printf("sources: %v, relSources: %v\n", sources, relSources)

	// fsmodtime.Logger = Logger

	// Paths tested with fs.FS can't start with "/". See https://pkg.go.dev/io/fs#ValidPath
	files, modified, err := fsmodtime.Target(os.DirFS("/"),
		[]string{filepath.Join("./", cwd, planFile)},
		filteredSources)
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", planFile)
	}
	if !ignoreCache && !modified {
		Logger.Printf("no changes: skipping plan\n")
		return nil
	}
	if len(files) > 0 {
		modifiedFiles := []string{}
		for _, f := range files {
			rel, err := filepath.Rel(cwd, "/"+f)
			if err != nil {
				rel = f
			}
			modifiedFiles = append(modifiedFiles, rel)
		}
		Logger.Printf("modified: %v\n", modifiedFiles)
	} else {
		Logger.Printf("missing target: %v\n", planFile)
	}

	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "plan", "-out", planFile}
	for _, v := range defaultVarFiles {
		cmd = append(cmd, "-var-file", v)
	}
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
	}
	for _, v := range variables {
		cmd = append(cmd, "-var", v)
	}
	if destroy {
		cmd = append(cmd, "-destroy")
	}
	if detailedExitcode {
		cmd = append(cmd, "-detailed-exitcode")
	}
	for _, t := range targets {
		cmd = append(cmd, "-target", t)
	}
	for _, r := range replacements {
		cmd = append(cmd, "-replace", r)
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd = append(cmd, "-no-color")
	}
	cmd = append(cmd, args...)

	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun)
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}
	err = ri.Run()
	if err != nil {
		// exit code 2 with detailed-exitcode means changes found
		var eerr *exec.ExitError
		if detailedExitcode && errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			Logger.Printf("plan has changes\n")
			return eerr
		}
		os.Remove(planFile)
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}
