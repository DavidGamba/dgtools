package main

import (
	"context"
	"fmt"

	esv1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1"
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// ExternalSecretWrapper provides version-agnostic access to ExternalSecret fields
type ExternalSecretWrapper struct {
	Name           string
	SecretStoreRef SecretStoreRef
	Status         StatusConditions
	Spec           interface{}
}

type SecretStoreRef struct {
	Name string
	Kind string
}

type StatusConditions struct {
	Reason string
}

// SecretStoreSpecWrapper provides version-agnostic access to SecretStore spec
type SecretStoreSpecWrapper struct {
	Spec interface{}
}

func ESChain(ctx context.Context, config *rest.Config, name, namespace, version string) error {
	// Get ExternalSecret using dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Define the GVR for ExternalSecret
	externalSecretGVR := schema.GroupVersionResource{
		Group:    "external-secrets.io",
		Version:  version,
		Resource: "externalsecrets",
	}

	// Get the ExternalSecret
	unstructuredES, err := dynamicClient.Resource(externalSecretGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ExternalSecret: %w", err)
	}

	// Convert to version-specific type
	esWrapper, err := convertExternalSecret(unstructuredES, version)
	if err != nil {
		return fmt.Errorf("failed to convert ExternalSecret: %w", err)
	}

	// Marshal ExternalSecret to YAML
	yamlData, err := marshalTrim(esWrapper.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal ExternalSecret to YAML: %w", err)
	}

	fmt.Printf("\n--- ExternalSecret: %s ---\n", esWrapper.Name)
	fmt.Printf("%s\n", string(yamlData))
	if esWrapper.Status.Reason != "" {
		fmt.Printf("Status: %s\n", esWrapper.Status.Reason)
	}

	// Get SecretStore using externalSecret.Spec.SecretStoreRef.Name
	secretStoreRefName := esWrapper.SecretStoreRef.Name
	secretStoreRefKind := esWrapper.SecretStoreRef.Kind

	// Determine the resource based on kind (SecretStore or ClusterSecretStore)
	var secretStoreGVR schema.GroupVersionResource
	var secretStoreNamespace string

	if secretStoreRefKind == "ClusterSecretStore" {
		secretStoreGVR = schema.GroupVersionResource{
			Group:    "external-secrets.io",
			Version:  version,
			Resource: "clustersecretstores",
		}
		secretStoreNamespace = "" // ClusterSecretStore is cluster-scoped
	} else {
		// Default to SecretStore
		secretStoreGVR = schema.GroupVersionResource{
			Group:    "external-secrets.io",
			Version:  version,
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

	// Convert to version-specific type
	ssWrapper, err := convertSecretStore(unstructuredSS, version, secretStoreRefKind)
	if err != nil {
		return fmt.Errorf("failed to convert %s: %w", secretStoreRefKind, err)
	}

	// Marshal SecretStore spec to YAML
	secretStoreYAML, err := marshalTrim(ssWrapper.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal %s to YAML: %w", secretStoreRefKind, err)
	}

	fmt.Printf("\n--- %s: %s ---\n", secretStoreRefKind, secretStoreRefName)
	fmt.Printf("%s\n", string(secretStoreYAML))
	return nil
}

// convertExternalSecret converts unstructured ExternalSecret to version-specific wrapper
func convertExternalSecret(unstructuredES *unstructured.Unstructured, version string) (*ExternalSecretWrapper, error) {
	wrapper := &ExternalSecretWrapper{}

	switch version {
	case "v1":
		es := &esv1.ExternalSecret{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			unstructuredES.UnstructuredContent(),
			es,
		)
		if err != nil {
			return nil, err
		}
		wrapper.Name = es.Name
		wrapper.SecretStoreRef = SecretStoreRef{
			Name: es.Spec.SecretStoreRef.Name,
			Kind: es.Spec.SecretStoreRef.Kind,
		}
		wrapper.Spec = es.Spec
		if len(es.Status.Conditions) > 0 {
			wrapper.Status = StatusConditions{
				Reason: string(es.Status.Conditions[0].Reason),
			}
		}

	case "v1beta1":
		es := &esv1beta1.ExternalSecret{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			unstructuredES.UnstructuredContent(),
			es,
		)
		if err != nil {
			return nil, err
		}
		wrapper.Name = es.Name
		wrapper.SecretStoreRef = SecretStoreRef{
			Name: es.Spec.SecretStoreRef.Name,
			Kind: es.Spec.SecretStoreRef.Kind,
		}
		wrapper.Spec = es.Spec
		if len(es.Status.Conditions) > 0 {
			wrapper.Status = StatusConditions{
				Reason: string(es.Status.Conditions[0].Reason),
			}
		}

	default:
		return nil, fmt.Errorf("unsupported version: %s", version)
	}

	return wrapper, nil
}

// convertSecretStore converts unstructured SecretStore/ClusterSecretStore to version-specific wrapper
func convertSecretStore(unstructuredSS *unstructured.Unstructured, version, kind string) (*SecretStoreSpecWrapper, error) {
	wrapper := &SecretStoreSpecWrapper{}

	switch version {
	case "v1":
		if kind == "ClusterSecretStore" {
			css := &esv1.ClusterSecretStore{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredSS.UnstructuredContent(),
				css,
			)
			if err != nil {
				return nil, err
			}
			wrapper.Spec = &css.Spec
		} else {
			ss := &esv1.SecretStore{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredSS.UnstructuredContent(),
				ss,
			)
			if err != nil {
				return nil, err
			}
			wrapper.Spec = &ss.Spec
		}

	case "v1beta1":
		if kind == "ClusterSecretStore" {
			css := &esv1beta1.ClusterSecretStore{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredSS.UnstructuredContent(),
				css,
			)
			if err != nil {
				return nil, err
			}
			wrapper.Spec = &css.Spec
		} else {
			ss := &esv1beta1.SecretStore{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredSS.UnstructuredContent(),
				ss,
			)
			if err != nil {
				return nil, err
			}
			wrapper.Spec = &ss.Spec
		}

	default:
		return nil, fmt.Errorf("unsupported version: %s", version)
	}

	return wrapper, nil
}
