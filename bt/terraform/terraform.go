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
	opt := parent.NewCommand("terraform", "terraform related tasks")
	opt.String("profile", "default", opt.Description("BT Terraform Profile to use"), opt.GetEnv("BT_TERRAFORM_PROFILE"))

	// backend-config
	initCMD(ctx, opt)

	// var-file
	planCMD(ctx, opt)
	importCMD(ctx, opt)
	refreshCMD(ctx, opt)

	// workspace selection
	applyCMD(ctx, opt)
	consoleCMD(ctx, opt)
	forceUnlockCMD(ctx, opt)
	outputCMD(ctx, opt)
	showCMD(ctx, opt)
	showPlanCMD(ctx, opt)
	stateListCMD(ctx, opt)
	statePullCMD(ctx, opt)
	statePushCMD(ctx, opt)
	stateRMCMD(ctx, opt)
	stateShowCMD(ctx, opt)
	taintCMD(ctx, opt)
	untaintCMD(ctx, opt)

	buildCMD(ctx, opt)

	return opt
}

// Retrieves workspaces assuming a convention where the .tfvars[.json] file matches the name of the workspace
// It only lists files, it doesn't query Terraform for a 'proper' list of workspaces.
func getWorkspaces(cfg *config.Config, profile string) ([]string, error) {
	wss := []string{}
	glob := fmt.Sprintf("%s/*.tfvars*", cfg.Terraform[profile].Workspaces.Dir)
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
	if cfg.Terraform[profile].Workspaces.Enabled {
		if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
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

func workspaceSelected() bool {
	if _, err := os.Stat(".terraform/environment"); os.IsNotExist(err) {
		return false
	}
	return true
}

func updateWSIfSelected(ws string) (string, error) {
	if workspaceSelected() {
		e, err := os.ReadFile(".terraform/environment")
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
	if cfg.Terraform[profile].Workspaces.Enabled {
		if !workspaceSelected() {
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
func AddVarFileIfWorkspaceSelected(cfg *config.Config, profile, ws string, varFiles []string) ([]string, error) {
	if ws != "" {
		glob := fmt.Sprintf("%s/%s.tfvars*", cfg.Terraform[profile].Workspaces.Dir, ws)
		Logger.Printf("ws: %s, glob: %s\n", ws, glob)
		ff, _, err := fsmodtime.Glob(os.DirFS("."), true, []string{glob})
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
	for _, vars := range cfg.Terraform[profile].Plan.VarFile {
		v := strings.ReplaceAll(vars, "~", "$HOME")
		vv, err := fsmodtime.ExpandEnv([]string{v})
		if err != nil {
			return varFiles, fmt.Errorf("failed to expand: %w", err)
		}
		if _, err := os.Stat(vv[0]); err == nil {
			varFiles = append(varFiles, vv[0])
		}
	}
	return varFiles, nil
}
