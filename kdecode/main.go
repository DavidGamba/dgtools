package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/DavidGamba/go-getoptions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
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
		var printer printers.ResourcePrinter
		switch output {
		case "yaml":
			printer = &printers.YAMLPrinter{}
		case "json":
			printer = &printers.JSONPrinter{}
		default:
			return fmt.Errorf("unknown output format: %s", output)
		}
		printer = &printers.OmitManagedFieldsPrinter{Delegate: printer}
		p := printers.NewTypeSetter(scheme.Scheme).ToPrinter(printer)
		err := p.PrintObj(k8sSecret, os.Stdout)
		if err != nil {
			return err
		}
		fmt.Println()
		return nil
	}

	// sort output
	dataKeys := []string{}
	for k := range k8sSecret.Data {
		dataKeys = append(dataKeys, k)
	}
	slices.Sort(dataKeys)
	stringKeys := []string{}
	for k := range k8sSecret.StringData {
		stringKeys = append(stringKeys, k)
	}
	slices.Sort(stringKeys)

	for _, k := range dataKeys {
		v := k8sSecret.Data[k]
		if (key != "" && k == key) || key == "" {
			fmt.Printf("%s=", k)
			if pem {
				info, err := ParseCert(string(v))
				if err != nil {
					fmt.Printf("%s\n", string(v))
					Logger.Printf("%s\n", err)
				} else {
					fmt.Printf("\n%s\n", info)
				}
			} else {
				fmt.Printf("%s\n", string(v))
			}
		}
	}

	for _, k := range stringKeys {
		v := k8sSecret.StringData[k]
		if (key != "" && k == key) || key == "" {
			fmt.Printf("%s=", k)
			if pem {
				info, err := ParseCert(v)
				if err != nil {
					fmt.Printf("%s\n", v)
					Logger.Printf("%s\n", err)
				} else {
					fmt.Printf("\n%s\n", info)
				}
			} else {
				fmt.Printf("%s\n", v)
			}
		}
	}

	return nil
}
