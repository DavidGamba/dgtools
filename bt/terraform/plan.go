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
	profile := parent.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("plan", "")
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("destroy", false)
	opt.Bool("detailed-exitcode", false)
	opt.Bool("ignore-cache", false, opt.Description("ignore the cache and re-run the plan"), opt.Alias("ic"))
	opt.StringSlice("target", 1, 99)
	opt.StringSlice("replace", 1, 99)
	opt.SetCommandFn(planRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

func planRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	destroy := opt.Value("destroy").(bool)
	detailedExitcode := opt.Value("detailed-exitcode").(bool)
	ignoreCache := opt.Value("ignore-cache").(bool)
	varFiles := opt.Value("var-file").([]string)
	targets := opt.Value("target").([]string)
	replacements := opt.Value("replace").([]string)
	ws := opt.Value("ws").(string)
	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	ws, err = getWorkspace(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	defaultVarFiles, err := getDefaultVarFiles(cfg, profile)
	if err != nil {
		return err
	}

	varFiles, err = AddVarFileIfWorkspaceSelected(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	planFile := ""
	if ws == "" {
		planFile = ".tf.plan"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
	}

	cwd, err := os.Getwd()
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

	cmd := []string{cfg.TFProfile[profile].BinaryName, "plan", "-out", planFile}
	for _, v := range defaultVarFiles {
		cmd = append(cmd, "-var-file", v)
	}
	for _, v := range varFiles {
		cmd = append(cmd, "-var-file", v)
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

	dataDir := fmt.Sprintf("TF_DATA_DIR=.terraform-%s", profile)
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log().Env(dataDir)
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

func checksRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	varFiles := opt.Value("var-file").([]string)
	ws := opt.Value("ws").(string)
	nc := opt.Value("no-checks").(bool)
	if nc {
		Logger.Printf("WARNING: no-checks flag passed. Skipping pre-apply checks.\n")
		return nil
	}

	ws, err := updateWSIfSelected(ws)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}

	cfg := config.ConfigFromContext(ctx)

	ws, err = getWorkspace(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	planFile := ""
	checkFile := ""
	if ws == "" {
		planFile = ".tf.plan"
		checkFile = ".tf.check"
	} else {
		planFile = fmt.Sprintf(".tf.plan-%s", ws)
		checkFile = fmt.Sprintf(".tf.check-%s", ws)
	}
	jsonPlan := planFile + ".json"
	os.Setenv("TERRAFORM_JSON_PLAN", jsonPlan)
	os.Setenv("CONFIG_ROOT", cfg.ConfigRoot)

	cmdFiles := []string{}
	for _, cmd := range cfg.TFProfile[profile].PreApplyChecks.Commands {
		exp, err := fsmodtime.ExpandEnv(cmd.Files)
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		for _, f := range exp {
			if strings.HasPrefix(f, "/") {
				cmdFiles = append(cmdFiles, filepath.Join("./", f))
			} else {
				cmdFiles = append(cmdFiles, filepath.Join("./", cwd, f))
			}
		}
	}
	globs, _, err := fsmodtime.Glob(os.DirFS("/"), false, cmdFiles)
	if err != nil {
		return fmt.Errorf("failed to glob sources: %w", err)
	}

	// Paths tested with fs.FS can't start with "/". See https://pkg.go.dev/io/fs#ValidPath
	files, modified, err := fsmodtime.Target(os.DirFS("/"),
		[]string{filepath.Join("./", cwd, checkFile)},
		append(globs, filepath.Join("./", cwd, planFile)))
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", jsonPlan)
	}
	Logger.Printf("plan in json format: %v\n", jsonPlan)

	if !modified {
		Logger.Printf("no changes: skipping check\n")
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

	cmd := []string{cfg.TFProfile[profile].BinaryName, "show", "-json", planFile}
	dataDir := fmt.Sprintf("TF_DATA_DIR=.terraform-%s", profile)
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log().Env(dataDir)
	out, err := ri.STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed to get plan json output: %w", err)
	}

	err = os.WriteFile(jsonPlan, out, 0600)
	if err != nil {
		return fmt.Errorf("failed to write json plan: %w", err)
	}
	Logger.Printf("plan json written to: %s\n", jsonPlan)

	for _, cmd := range cfg.TFProfile[profile].PreApplyChecks.Commands {
		Logger.Printf("running check: %s\n", cmd.Name)
		exp, err := fsmodtime.ExpandEnv(cmd.Command)
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		ri := run.CMD(exp...).Ctx(ctx).Stdin().Log().Env(dataDir)
		err = ri.Run()
		if err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}
	}

	fh, err := os.Create(checkFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()

	return nil
}
