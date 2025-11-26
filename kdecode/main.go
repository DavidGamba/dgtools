package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Self("", `Decodes the given secret.
If a secret is not given, lists all secrets in the current namespace.

Source: https://github.com/DavidGamba/dgtools`)
	opt.SetCommandFn(Run)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.Bool("pem", false, opt.Description("Parse the secret as a PEM encoded certificate"))
	opt.Bool("es", false, opt.Description("Show External Secrets Operator chain"))
	opt.String("namespace", "")
	opt.String("key", "", opt.Description("limit output to the given key"))
	opt.String("cluster", "", opt.Description(`Allows targeting a different cluster from the current context.

NOTE: This only works when you are not in a subshell
      that sets a KUBECONFIG subset like 'kubie'.
`))
	opt.String("output", "", opt.Description("Provide an output format to get the raw YAML/JSON"), opt.ValidValues("yaml", "json"), opt.ArgName("yaml|json"))
	opt.HelpSynopsisArg("[<secret>]", "secret name")
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	namespace := opt.Value("namespace").(string)
	cluster := opt.Value("cluster").(string)
	key := opt.Value("key").(string)
	pem := opt.Value("pem").(bool)
	output := opt.Value("output").(string)
	es := opt.Value("es").(bool)
	secret := ""
	if len(args) >= 1 {
		secret = args[0]
	}

	kubeConfig := GetKubeConfig()
	Logger.Printf("kubeConfig: %s", kubeConfig)
	config, err := NewRestConfig(kubeConfig, cluster)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}
	if namespace == "" {
		namespace, _, _ = GetNamespace(kubeConfig, cluster)
		Logger.Printf("namespace: %s", namespace)
	}
	if secret == "" {
		list, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}
		for _, s := range list.Items {
			fmt.Println(s.Name)
		}
		return nil
	}
	k8sSecret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, secret, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if output != "" {
		return resourcePrint(output, k8sSecret, os.Stdout)
	}
	SecretPrint(k8sSecret, pem, key)

	if !es {
		return nil
	}
	if len(k8sSecret.OwnerReferences) > 0 {
		ownerRef := k8sSecret.OwnerReferences[0]
		if ownerRef.Kind == "ExternalSecret" {
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
				Get(ctx, ownerRef.Name, metav1.GetOptions{})
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
		}
	}

	return nil
}
