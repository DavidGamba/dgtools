package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/trees"
	"github.com/DavidGamba/dgtools/yamlutils"
)

func GetK8sContext(ctx context.Context) (contextName, namespace string, err error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	ymlList, err := yamlutils.NewFromFile(kubeconfig)
	if err != nil {
		err = fmt.Errorf("failed to get kube config: %w", err)
		return
	}
	if len(ymlList) < 1 {
		err = fmt.Errorf("no resources found in kubeconfig")
		return
	}
	current, err := ymlList[0].GetString(false, []string{"current-context"})
	if err != nil {
		err = fmt.Errorf("failed to get current context: %w", err)
		return
	}
	contextName = current

	kctxsI, _, err := trees.NavigateTree(false, ymlList[0].Tree, []string{"contexts"})
	if err != nil {
		err = fmt.Errorf("failed to get contexts: %w", err)
		return
	}
	kctxs, ok := kctxsI.([]any)
	if !ok {
		err = fmt.Errorf("failed to convert contexts: %w", err)
		return
	}

	for _, kctx := range kctxs {
		name := kctx.(map[string]any)["name"].(string)
		if name != current {
			continue
		}
		c := kctx.(map[string]any)["context"]
		ns := c.(map[string]any)["namespace"]
		if ns != nil {
			namespace = ns.(string)
		}
	}

	return contextName, namespace, nil
}
