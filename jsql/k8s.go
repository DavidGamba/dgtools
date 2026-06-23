package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidGamba/dgtools/run"
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

func GetK8sResource(ctx context.Context, cacheDir, resource string) error {
	cmd := []string{"kubectl", "get", "-A", "-o", "json", resource}
	out, err := run.CMD(cmd...).Log().STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	cmd = []string{"qq", ".items", "-o", "json"}
	out, err = run.CMD(cmd...).In(out).STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	filename := filepath.Join(cacheDir, resource+".json")
	fh, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fh.Close()
	_, err = fh.Write(out)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func UpdateK8sResourceQueries(cacheDir, resource string) []string {
	filename := filepath.Join(cacheDir, resource+".json")
	cmds := []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s;", resource),
		fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM '%s';", resource, filename),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN name VARCHAR;", resource),
		fmt.Sprintf("UPDATE %s SET name = metadata.name;", resource),
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN namespace VARCHAR;", resource),
		// Use cast to allow for null values
		fmt.Sprintf("UPDATE %s SET namespace = CAST(metadata AS JSON)->>'namespace';", resource),
	}
	return cmds
}
