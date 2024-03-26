package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	opt.String("namespace", "")
	opt.String("key", "", opt.Description("limit output to the given key"))
	opt.String("cluster", "", opt.Description(`Allows targeting a different cluster from the current context.

NOTE: This only works when you are not in a subshell
      that sets a KUBECONFIG subset like 'kubie'.
`))
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
	for k, v := range k8sSecret.Data {
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
	for k, v := range k8sSecret.StringData {
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

func ParseCert(data string) (string, error) {
	info := ""
	blocks := []*pem.Block{}
	rest := []byte(data)
	for {
		block, newRest := pem.Decode(rest)
		if block == nil {
			return info, fmt.Errorf("failed to parse certificate PEM")
		}
		blocks = append(blocks, block)
		if len(newRest) == 0 {
			break
		}
		rest = newRest
	}

	for i, block := range blocks {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return info, fmt.Errorf("failed to parse certificate: %w", err)
		}
		info += fmt.Sprintf("Certificate %d\n", i)
		info += fmt.Sprintf("\tIssuer: %s\n", cert.Issuer)
		info += fmt.Sprintf("\tSubject: %s\n", cert.Subject)
		info += fmt.Sprintf("\tNotBefore: %s\n", cert.NotBefore)
		info += fmt.Sprintf("\tNotAfter: %s\n", cert.NotAfter)
		info += fmt.Sprintf("\tSignatureAlgorithm: %s\n", cert.SignatureAlgorithm)
		info += fmt.Sprintf("\tPublicKeyAlgorithm: %s\n", cert.PublicKeyAlgorithm)
		info += fmt.Sprintf("\tSerialNumber: %s\n", cert.SerialNumber)
		info += fmt.Sprintf("\tKeyUsage: %s\n", KeyUsage(cert.KeyUsage))
		if len(cert.OCSPServer) > 0 {
			info += fmt.Sprintf("\tOCSPServer: %v\n", cert.OCSPServer)
		}
		if len(cert.IssuingCertificateURL) > 0 {
			info += fmt.Sprintf("\tIssuingCertificateURL: %v\n", cert.IssuingCertificateURL)
		}
		if len(cert.DNSNames) > 0 {
			info += fmt.Sprintf("\tDNSNames: %v\n", cert.DNSNames)
		}
		if len(cert.EmailAddresses) > 0 {
			info += fmt.Sprintf("\tEmailAddresses: %v\n", cert.EmailAddresses)
		}
		if len(cert.IPAddresses) > 0 {
			info += fmt.Sprintf("\tIPAddresses: %v\n", cert.IPAddresses)
		}
		if len(cert.URIs) > 0 {
			info += fmt.Sprintf("\tURIs: %v\n", cert.URIs)
		}
		if len(cert.CRLDistributionPoints) > 0 {
			info += fmt.Sprintf("\tCRLDistributionPoints: %v\n", cert.CRLDistributionPoints)
		}
	}
	return info, nil
}

func KeyUsage(k x509.KeyUsage) string {
	switch k {
	case x509.KeyUsageDigitalSignature:
		return "DigitalSignature"
	case x509.KeyUsageContentCommitment:
		return "ContentCommitment"
	case x509.KeyUsageKeyEncipherment:
		return "KeyEncipherment"
	case x509.KeyUsageDataEncipherment:
		return "DataEncipherment"
	case x509.KeyUsageKeyAgreement:
		return "KeyAgreement"
	case x509.KeyUsageCertSign:
		return "CertSign"
	case x509.KeyUsageCRLSign:
		return "CRLSign"
	case x509.KeyUsageEncipherOnly:
		return "EncipherOnly"
	case x509.KeyUsageDecipherOnly:
		return "DecipherOnly"
	}
	return ""
}
