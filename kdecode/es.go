package main

import (
	"context"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func ESChain(ctx context.Context, config *rest.Config, name, namespace string) error {
	// Get ExternalSecret using dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Define the GVR for ExternalSecret
	externalSecretGVR := schema.GroupVersionResource{
		Group:    "external-secrets.io",
		Version:  "v1beta1",
		Resource: "externalsecrets",
	}

	// Get the ExternalSecret
	unstructuredES, err := dynamicClient.Resource(externalSecretGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ExternalSecret: %w", err)
	}

	// Convert unstructured to typed ExternalSecret
	externalSecret := &esv1beta1.ExternalSecret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		unstructuredES.UnstructuredContent(),
		externalSecret,
	)
	if err != nil {
		return fmt.Errorf("failed to convert ExternalSecret: %w", err)
	}

	// Marshal ExternalSecret to YAML
	yamlData, err := marshalTrim(externalSecret.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal ExternalSecret to YAML: %w", err)
	}

	fmt.Printf("\n--- ExternalSecret: %s ---\n", externalSecret.Name)
	fmt.Printf("%s\n", string(yamlData))
	if len(externalSecret.Status.Conditions) > 0 {
		fmt.Printf("Status: %s\n", externalSecret.Status.Conditions[0].Reason)
	}

	// Get SecretStore using externalSecret.Spec.SecretStoreRef.Name
	secretStoreRefName := externalSecret.Spec.SecretStoreRef.Name
	secretStoreRefKind := externalSecret.Spec.SecretStoreRef.Kind

	// Determine the resource based on kind (SecretStore or ClusterSecretStore)
	var secretStoreGVR schema.GroupVersionResource
	var secretStoreNamespace string

	if secretStoreRefKind == "ClusterSecretStore" {
		secretStoreGVR = schema.GroupVersionResource{
			Group:    "external-secrets.io",
			Version:  "v1beta1",
			Resource: "clustersecretstores",
		}
		secretStoreNamespace = "" // ClusterSecretStore is cluster-scoped
	} else {
		// Default to SecretStore
		secretStoreGVR = schema.GroupVersionResource{
			Group:    "external-secrets.io",
			Version:  "v1beta1",
			Resource: "secretstores",
		}
		secretStoreNamespace = namespace
	}

	// Get the SecretStore or ClusterSecretStore
	var unstructuredSS *unstructured.Unstructured
	var secretStoreErr error

	if secretStoreNamespace == "" {
		// ClusterSecretStore - no namespace
		unstructuredSS, secretStoreErr = dynamicClient.Resource(secretStoreGVR).
			Get(ctx, secretStoreRefName, metav1.GetOptions{})
	} else {
		// SecretStore - namespaced
		unstructuredSS, secretStoreErr = dynamicClient.Resource(secretStoreGVR).
			Namespace(secretStoreNamespace).
			Get(ctx, secretStoreRefName, metav1.GetOptions{})
	}

	if secretStoreErr != nil {
		return fmt.Errorf("failed to get %s: %w", secretStoreRefKind, secretStoreErr)
	}

	// Convert unstructured to typed SecretStore or ClusterSecretStore to access spec
	var secretStoreSpec *esv1beta1.SecretStoreSpec

	if secretStoreRefKind == "ClusterSecretStore" {
		clusterSecretStore := &esv1beta1.ClusterSecretStore{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(
			unstructuredSS.UnstructuredContent(),
			clusterSecretStore,
		)
		if err != nil {
			return fmt.Errorf("failed to convert ClusterSecretStore: %w", err)
		}
		secretStoreSpec = &clusterSecretStore.Spec
	} else {
		secretStore := &esv1beta1.SecretStore{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(
			unstructuredSS.UnstructuredContent(),
			secretStore,
		)
		if err != nil {
			return fmt.Errorf("failed to convert SecretStore: %w", err)
		}
		secretStoreSpec = &secretStore.Spec
	}

	// Marshal SecretStore spec to YAML
	secretStoreYAML, err := marshalTrim(secretStoreSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal %s to YAML: %w", secretStoreRefKind, err)
	}

	fmt.Printf("\n--- %s: %s ---\n", secretStoreRefKind, secretStoreRefName)
	fmt.Printf("%s\n", string(secretStoreYAML))
	return nil
}
