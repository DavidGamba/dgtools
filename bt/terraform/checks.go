package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
)

func checksCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("checks", "Run pre-apply checks against the latest plan")
	opt.Bool("dry-run", false)
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("no-checks", false, opt.Description("Do not run pre-apply checks"), opt.Alias("nc"))
	opt.Bool("ignore-cache", false, opt.Description("ignore the cache and re-run the checks"), opt.Alias("ic"))
	opt.SetCommandFn(checksRun)

	return opt
}

func checksRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	varFiles := opt.Value("var-file").([]string)
	ws := opt.Value("ws").(string)
	ignoreCache := opt.Value("ignore-cache").(bool)
	nc := opt.Value("no-checks").(bool)
	if nc {
		Logger.Printf("WARNING: no-checks flag passed. Skipping pre-apply checks.\n")
		return nil
	}

	cfg := config.ConfigFromContext(ctx)
	component := ComponentFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	cwd, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}
	if component == "." {
		component = filepath.Base(cwd)
	}
	component = strings.Split(component, ":")[0]
	Logger.Printf("component: %s\n", component)

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
	txtPlan := planFile + ".txt"
	wsEnv := ws
	if ws == "" {
		wsEnv = "default"
	}

	env := map[string]string{
		"TERRAFORM_JSON_PLAN": jsonPlan,
		"TERRAFORM_TXT_PLAN":  txtPlan,
		"CONFIG_ROOT":         cfg.ConfigRoot,
		"TF_WORKSPACE":        wsEnv,
		"BT_COMPONENT":        component,
	}

	cmdFiles := []string{}
	for _, cmd := range cfg.TFProfile[cfg.Profile(profile)].PreApplyChecks.Commands {
		exp, err := fsmodtime.ExpandEnv(cmd.Files, env)
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

	if !ignoreCache && !modified {
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
		Logger.Printf("missing target: %v\n", checkFile)
	}

	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)
	cmd := []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "show", "-json", planFile}
	ri := run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun)
	out, err := ri.STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed to get plan json output: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, jsonPlan), out, 0600)
	if err != nil {
		return fmt.Errorf("failed to write json plan: %w", err)
	}
	Logger.Printf("plan json written to: %s\n", jsonPlan)

	cmd = []string{cfg.TFProfile[cfg.Profile(profile)].BinaryName, "show", planFile}
	ri = run.CMDCtx(ctx, cmd...).Stdin().Log().Env(dataDir).Dir(dir).DryRun(dryRun)
	out, err = ri.STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed to get plan json output: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, txtPlan), out, 0600)
	if err != nil {
		return fmt.Errorf("failed to write txt plan: %w", err)
	}
	Logger.Printf("plan txt written to: %s\n", txtPlan)

	for _, cmd := range cfg.TFProfile[cfg.Profile(profile)].PreApplyChecks.Commands {
		Logger.Printf("running check: %s\n", cmd.Name)
		exp, err := fsmodtime.ExpandEnv(cmd.Command, env)
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		ri := run.CMDCtx(ctx, exp...).Stdin().Log().
			Env(dataDir).
			Env(fmt.Sprintf("TERRAFORM_JSON_PLAN=%s", jsonPlan)).
			Env(fmt.Sprintf("TERRAFORM_TXT_PLAN=%s", txtPlan)).
			Env(fmt.Sprintf("CONFIG_ROOT=%s", cfg.ConfigRoot)).
			Env(fmt.Sprintf("TF_WORKSPACE=%s", wsEnv)).
			Env(fmt.Sprintf("BT_COMPONENT=%s", component)).
			Dir(dir).DryRun(dryRun)
		if cmd.OutputFile == "" {
			err = ri.Run()
			if err != nil {
				return fmt.Errorf("failed to run: %w", err)
			}
		} else {
			f, err := fsmodtime.ExpandEnv([]string{cmd.OutputFile}, env)
			if err != nil {
				return fmt.Errorf("failed to expand: %w", err)
			}
			fh, err := os.Create(filepath.Join(dir, f[0]))
			if err != nil {
				return fmt.Errorf("failed to create cmd output file: %w", err)
			}
			defer fh.Close()
			err = ri.Run(fh, os.Stderr)
			if err != nil {
				return fmt.Errorf("failed to run: %w", err)
			}
			Logger.Printf("Saving check output to file: %s\n", filepath.Join(dir, f[0]))
		}
	}

	if dryRun {
		return nil
	}

	fh, err := os.Create(checkFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()

	return nil
}

func postChecksCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	opt := parent.NewCommand("post-checks", "Run post-apply checks")
	opt.Bool("dry-run", false)
	opt.StringSlice("var-file", 1, 1)
	opt.Bool("no-checks", false, opt.Description("Do not run post-apply checks"), opt.Alias("nc"))
	opt.Bool("ignore-cache", false, opt.Description("ignore the cache and re-run the checks"), opt.Alias("ic"))
	opt.SetCommandFn(postChecksRun)

	return opt
}

func postChecksRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	dryRun := opt.Value("dry-run").(bool)
	profile := opt.Value("profile").(string)
	varFiles := opt.Value("var-file").([]string)
	ws := opt.Value("ws").(string)
	ignoreCache := opt.Value("ignore-cache").(bool)
	nc := opt.Value("no-checks").(bool)
	if nc {
		Logger.Printf("WARNING: no-checks flag passed. Skipping post-apply checks.\n")
		return nil
	}

	cfg := config.ConfigFromContext(ctx)
	component := ComponentFromContext(ctx)
	dir := DirFromContext(ctx)
	LogConfig(cfg, profile)

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile), ws)
	if err != nil {
		return err
	}

	cwd, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}
	if component == "." {
		component = filepath.Base(cwd)
	}
	component = strings.Split(component, ":")[0]
	Logger.Printf("component: %s\n", component)

	ws, err = getWorkspace(cfg, profile, ws, varFiles)
	if err != nil {
		return err
	}

	checkFile := ""
	if ws == "" {
		checkFile = ".tf.postcheck"
	} else {
		checkFile = fmt.Sprintf(".tf.postcheck-%s", ws)
	}
	wsEnv := ws
	if ws == "" {
		wsEnv = "default"
	}

	env := map[string]string{
		"CONFIG_ROOT":  cfg.ConfigRoot,
		"TF_WORKSPACE": wsEnv,
		"BT_COMPONENT": component,
	}

	cmdFiles := []string{}
	for _, cmd := range cfg.TFProfile[cfg.Profile(profile)].PostApplyChecks.Commands {
		exp, err := fsmodtime.ExpandEnv(cmd.Files, env)
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
		globs)
	if err != nil {
		Logger.Printf("failed to check changes for: '%s'\n", checkFile)
	}

	if !ignoreCache && !modified {
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
		Logger.Printf("missing target: %v\n", checkFile)
	}

	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)))
	Logger.Printf("export %s\n", dataDir)

	for _, cmd := range cfg.TFProfile[cfg.Profile(profile)].PostApplyChecks.Commands {
		Logger.Printf("running check: %s\n", cmd.Name)
		exp, err := fsmodtime.ExpandEnv(cmd.Command, env)
		if err != nil {
			return fmt.Errorf("failed to expand: %w", err)
		}
		ri := run.CMDCtx(ctx, exp...).Stdin().Log().
			Env(dataDir).
			Env(fmt.Sprintf("CONFIG_ROOT=%s", cfg.ConfigRoot)).
			Env(fmt.Sprintf("TF_WORKSPACE=%s", wsEnv)).
			Env(fmt.Sprintf("BT_COMPONENT=%s", component)).
			Dir(dir).DryRun(dryRun)
		if cmd.OutputFile == "" {
			err = ri.Run()
			if err != nil {
				return fmt.Errorf("failed to run: %w", err)
			}
		} else {
			f, err := fsmodtime.ExpandEnv([]string{cmd.OutputFile}, env)
			if err != nil {
				return fmt.Errorf("failed to expand: %w", err)
			}
			fh, err := os.Create(filepath.Join(dir, f[0]))
			if err != nil {
				return fmt.Errorf("failed to create cmd output file: %w", err)
			}
			defer fh.Close()
			err = ri.Run(fh, os.Stderr)
			if err != nil {
				return fmt.Errorf("failed to run: %w", err)
			}
			Logger.Printf("Saving check output to file: %s\n", filepath.Join(dir, f[0]))
		}
	}

	if dryRun {
		return nil
	}

	fh, err := os.Create(checkFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	fh.Close()

	return nil
}
