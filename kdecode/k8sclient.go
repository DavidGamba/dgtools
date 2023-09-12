package main

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeConfig string

func GetKubeConfig() KubeConfig {
	// Allow for kubie and other tools that set a different KUBECONFIG
	c := os.Getenv("KUBECONFIG")
	if c == "" {
		homeDir, _ := os.UserHomeDir()
		c = filepath.Join(homeDir, ".kube", "config")
	}
	return KubeConfig(c)
}

func NewRestConfig(c KubeConfig, cluster string) (*rest.Config, error) {
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: string(c)}
	configOverrides := &clientcmd.ConfigOverrides{}
	if cluster != "" {
		configOverrides.CurrentContext = cluster
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config for cluster '%s': %w", cluster, err)
	}
	return config, nil
}

func GetNamespace(c KubeConfig, cluster string) (string, bool, error) {
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: string(c)}
	configOverrides := &clientcmd.ConfigOverrides{}
	if cluster != "" {
		configOverrides.CurrentContext = cluster
	}
	ns, overriden, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides).Namespace()
	if err != nil {
		return ns, overriden, fmt.Errorf("failed to get namespace for cluster '%s': %w", cluster, err)
	}
	return ns, overriden, nil
}
