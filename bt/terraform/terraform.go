package terraform

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/fsmodtime"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func NewCommand(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("terraform", "terraform related tasks")
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv(cfg.Config.TerraformProfileEnvVar))
	opt.Bool("tf-in-automation", false, opt.Description(`Determine if we are running in automation.
It will use a separate TF_DATA_DIR per workspace.`), opt.GetEnv("TF_IN_AUTOMATION"), opt.GetEnv("BT_IN_AUTOMATION"))

	opt.String("ws", "", opt.Description("Workspace to use"),
		opt.SuggestedValuesFn(func(target string, partialCompletion string) []string {
			wss, err := validWorkspaces(cfg, opt.Value("profile").(string))
			if err != nil {
				Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
			}
			return wss
		}),
	)

	// backend-config
	initCMD(ctx, opt)

	// var-file
	planCMD(ctx, opt)
	importCMD(ctx, opt)
	refreshCMD(ctx, opt)
	consoleCMD(ctx, opt)
	testCMD(ctx, opt)

	// workspace management
	workspaceCMD(ctx, opt)

	// workspace selection
	applyCMD(ctx, opt)
	forceUnlockCMD(ctx, opt)
	outputCMD(ctx, opt)
	showCMD(ctx, opt)
	showPlanCMD(ctx, opt)
	stateCMD(ctx, opt)
	taintCMD(ctx, opt)
	untaintCMD(ctx, opt)
	validateCMD(ctx, opt)
	providersCMD(ctx, opt)
	graphCMD(ctx, opt)

	// Custom
	buildCMD(ctx, opt)
	checksCMD(ctx, opt)
	postChecksCMD(ctx, opt)

	return opt
}

func LogConfig(cfg *config.Config, profile string) {
	Logger.Printf("cfg: %s\n", cfg.TFProfile[cfg.Profile(profile)])
	if cfg.ConfigFile == "" {
		Logger.Printf("WARNING: cfg file not found\n")
	}
}

// Retrieves workspaces assuming a convention where the .tfvars[.json] file matches the name of the workspace
// It only lists files, it doesn't query Terraform for a 'proper' list of workspaces.
func getWorkspaces(cfg *config.Config, profile string) ([]string, error) {
	wss := []string{}
	glob := fmt.Sprintf("%s/*.tfvars*", cfg.TFProfile[cfg.Profile(profile)].Workspaces.Dir)
	ff, _, err := fsmodtime.Glob(os.DirFS("."), true, []string{glob})
	if err != nil {
		return wss, fmt.Errorf("failed to glob ws files: %w", err)
	}
	for _, ws := range ff {
		ws = filepath.Base(ws)
		ws = strings.TrimSuffix(ws, ".json")
		ws = strings.TrimSuffix(ws, ".tfvars")
		wss = append(wss, ws)
	}
	return wss, nil
}

func validWorkspaces(cfg *config.Config, profile string) ([]string, error) {
	wss := []string{}
	if cfg.TFProfile[cfg.Profile(profile)].Workspaces.Enabled {
		envFile := getDataDir(cfg.Config.DefaultTerraformProfile, cfg.Profile(profile)) + "/environment"
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			wss, err = getWorkspaces(cfg, profile)
			if err != nil {
				return wss, err
			}
		} else {
			e, err := os.ReadFile(".terraform/environment")
			if err != nil {
				return wss, err
			}
			wss = append(wss, strings.TrimSpace(string(e)))
		}
	}
	return wss, nil
}

func getDataDir(defaultProfile, profile string) string {
	envFile := ".terraform"
	if defaultProfile != profile {
		envFile = fmt.Sprintf(".terraform-%s", profile)
	}
	return envFile
}

func workspaceSelected(defaultProfile, profile string) bool {
	envFile := getDataDir(defaultProfile, profile) + "/environment"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// If the given workspace is empty and there is a workspace selected then use the selected workspace
func updateWSIfSelected(defaultProfile, profile, ws string) (string, error) {
	if workspaceSelected(defaultProfile, profile) {
		envFile := getDataDir(defaultProfile, profile) + "/environment"
		e, err := os.ReadFile(envFile)
		if err != nil {
			return ws, fmt.Errorf("failed to read current workspace: %w", err)
		}
		wse := strings.TrimSpace(string(e))
		if ws != "" && wse != ws {
			return wse, fmt.Errorf("given workspace doesn't match selected workspace: %s", wse)
		}
		ws = wse
	}
	return ws, nil
}

// If there is no workspace selected, check the given var files and use the first one as the workspace then return the ws env var
func getWorkspace(cfg *config.Config, profile, ws string, varFiles []string) (string, error) {
	if cfg.TFProfile[cfg.Profile(profile)].Workspaces.Enabled {
		if !workspaceSelected(cfg.Config.DefaultTerraformProfile, profile) {
			if ws != "" {
				return ws, nil
			}
			if len(varFiles) < 1 {
				return "", fmt.Errorf("running in workspace mode but no workspace selected or -var-file given")
			}
			wsFilename := filepath.Base(varFiles[0])
			r := regexp.MustCompile(`\..*$`)
			ws = r.ReplaceAllString(wsFilename, "")
		}
	}
	return ws, nil
}

// If a workspace is selected automatically insert a var file matching the workspace.
// If the var file is already present then don't add it again.
func AddVarFileIfWorkspaceSelected(cfg *config.Config, profile, dir, ws string, varFiles []string) ([]string, error) {
	if ws != "" {
		glob := fmt.Sprintf("%s/%s.tfvars*", cfg.TFProfile[cfg.Profile(profile)].Workspaces.Dir, ws)
		Logger.Printf("ws: %s, glob: %s\n", ws, glob)
		ff, _, err := fsmodtime.Glob(os.DirFS(dir), true, []string{glob})
		if err != nil {
			return varFiles, fmt.Errorf("failed to glob ws files: %w", err)
		}
		for _, f := range ff {
			Logger.Printf("file: %s\n", f)
			if !slices.Contains(varFiles, f) {
				varFiles = append(varFiles, f)
			}
		}
	}
	return varFiles, nil
}

func getDefaultVarFiles(cfg *config.Config, profile string) ([]string, error) {
	varFiles := []string{}
	for _, vars := range cfg.TFProfile[cfg.Profile(profile)].Plan.VarFile {
		v := strings.ReplaceAll(vars, "~", "$HOME")
		vv, err := fsmodtime.ExpandEnv([]string{v}, nil)
		if err != nil {
			return varFiles, fmt.Errorf("failed to expand: %w", err)
		}
		// TODO: Consider re-introducing validation
		// if _, err := os.Stat(vv[0]); err == nil {
		// }
		varFiles = append(varFiles, vv[0])
	}
	return varFiles, nil
}
