package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
	"github.com/golang-jwt/jwt/v5"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))

	hmac := opt.NewCommand("hmac", "Create a JWT signed with an HMAC secret")
	hmac.StringMap("claim", 1, 99, opt.Alias("c"), opt.Description("Custom claim as key=value, can be given multiple times"))
	hmac.String("subject", "", opt.Alias("s"), opt.Description("Subject (sub) claim"))
	hmac.String("issuer", "", opt.Alias("i"), opt.Description("Issuer (iss) claim"))
	hmac.String("expires-in", "1h", opt.Alias("e"), opt.Description("Expiration duration from now (e.g. 1h, 30m)"))
	hmac.String("secret", "", opt.Alias("k"), opt.Required(), opt.GetEnv("JWT_SECRET"), opt.Description("HMAC secret used to sign the token"))
	hmac.SetCommandFn(HMACRun)

	rsaCmd := opt.NewCommand("rsa", "Create a JWT signed with an RSA private key from a PEM file")
	rsaCmd.StringMap("claim", 1, 99, opt.Alias("c"), opt.Description("Custom claim as key=value, can be given multiple times"))
	rsaCmd.String("subject", "", opt.Alias("s"), opt.Description("Subject (sub) claim"))
	rsaCmd.String("issuer", "", opt.Alias("i"), opt.Description("Issuer (iss) claim"))
	rsaCmd.String("expires-in", "1h", opt.Alias("e"), opt.Description("Expiration duration from now (e.g. 1h, 30m)"))
	rsaCmd.String("pem-file", "", opt.Alias("p"), opt.Required(), opt.Description("Path to the RSA private key PEM file used to sign the token"))
	rsaCmd.SetCommandFn(RSARun)

	ecCmd := opt.NewCommand("ec", "Create a JWT signed with an EC (P-256) private key from a PEM file (ES256)")
	ecCmd.StringMap("claim", 1, 99, opt.Alias("c"), opt.Description("Custom claim as key=value, can be given multiple times"))
	ecCmd.String("subject", "", opt.Alias("s"), opt.Description("Subject (sub) claim"))
	ecCmd.String("issuer", "", opt.Alias("i"), opt.Description("Issuer (iss) claim"))
	ecCmd.String("expires-in", "1h", opt.Alias("e"), opt.Description("Expiration duration from now (e.g. 1h, 30m)"))
	ecCmd.String("pem-file", "", opt.Alias("p"), opt.Required(), opt.Description("Path to the EC private key PEM file used to sign the token"))
	ecCmd.SetCommandFn(ECRun)

	verify := opt.NewCommand("verify", "Verify and decode a JWT")
	verify.String("secret", "", opt.Alias("k"), opt.GetEnv("JWT_SECRET"), opt.Description("HMAC secret used to verify the token (for HS256 tokens)"))
	verify.String("pem-file", "", opt.Alias("p"), opt.Description("Path to the RSA/EC public key or certificate PEM file used to verify the token (for RS256/ES256 tokens)"))
	verify.HelpSynopsisArg("<token>", "JWT token to verify")
	verify.SetCommandFn(VerifyRun)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		fmt.Fprintln(os.Stderr, opt.Help())
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}
	Logger.Println(remaining)

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

func HMACRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	secret := opt.Value("secret").(string)
	subject := opt.Value("subject").(string)
	issuer := opt.Value("issuer").(string)
	expiresInStr := opt.Value("expires-in").(string)
	claims := opt.Value("claim").(map[string]string)

	mapClaims, err := buildClaims(subject, issuer, expiresInStr, claims)
	if err != nil {
		return err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return fmt.Errorf("failed to sign token: %w", err)
	}

	fmt.Println(signed)
	return nil
}

func RSARun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	pemFile := opt.Value("pem-file").(string)
	subject := opt.Value("subject").(string)
	issuer := opt.Value("issuer").(string)
	expiresInStr := opt.Value("expires-in").(string)
	claims := opt.Value("claim").(map[string]string)

	mapClaims, err := buildClaims(subject, issuer, expiresInStr, claims)
	if err != nil {
		return err
	}

	key, err := loadRSAPrivateKey(pemFile)
	if err != nil {
		return err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)
	signed, err := token.SignedString(key)
	if err != nil {
		return fmt.Errorf("failed to sign token: %w", err)
	}

	fmt.Println(signed)
	return nil
}

func VerifyRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	tokenString, _, err := opt.GetRequiredArg(args)
	if err != nil {
		return err
	}
	secret := opt.Value("secret").(string)
	pemFile := opt.Value("pem-file").(string)

	keyFunc := func(token *jwt.Token) (any, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			if secret == "" {
				return nil, fmt.Errorf("token is signed with HMAC (%s) but no --secret|-k was provided", token.Header["alg"])
			}
			return []byte(secret), nil
		case *jwt.SigningMethodRSA:
			if pemFile == "" {
				return nil, fmt.Errorf("token is signed with RSA (%s) but no --pem-file|-p was provided", token.Header["alg"])
			}
			return loadRSAPublicKey(pemFile)
		case *jwt.SigningMethodECDSA:
			if pemFile == "" {
				return nil, fmt.Errorf("token is signed with EC (%s) but no --pem-file|-p was provided", token.Header["alg"])
			}
			return loadECPublicKey(pemFile)
		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
	}

	parsed, err := jwt.Parse(tokenString, keyFunc)
	if parsed != nil {
		if claims, ok := parsed.Claims.(jwt.MapClaims); ok {
			pretty, mErr := json.MarshalIndent(claims, "", "  ")
			if mErr == nil {
				fmt.Println(string(pretty))
			}
		}
	}
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}
	if !parsed.Valid {
		return fmt.Errorf("token is invalid")
	}

	fmt.Fprintln(os.Stderr, "token is valid")
	return nil
}

func loadRSAPublicKey(pemFile string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem file: no PEM block found")
	}

	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		rsaKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate does not contain an RSA public key")
		}
		return rsaKey, nil
	}

	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an RSA key")
	}
	return rsaKey, nil
}

func ECRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	pemFile := opt.Value("pem-file").(string)
	subject := opt.Value("subject").(string)
	issuer := opt.Value("issuer").(string)
	expiresInStr := opt.Value("expires-in").(string)
	claims := opt.Value("claim").(map[string]string)

	mapClaims, err := buildClaims(subject, issuer, expiresInStr, claims)
	if err != nil {
		return err
	}

	key, err := loadECPrivateKey(pemFile)
	if err != nil {
		return err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, mapClaims)
	signed, err := token.SignedString(key)
	if err != nil {
		return fmt.Errorf("failed to sign token: %w", err)
	}

	fmt.Println(signed)
	return nil
}

func loadECPrivateKey(pemFile string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem file: no PEM block found")
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	ecKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an EC key")
	}
	return ecKey, nil
}

func loadECPublicKey(pemFile string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem file: no PEM block found")
	}

	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		ecKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate does not contain an EC public key")
		}
		return ecKey, nil
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	ecKey, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an EC key")
	}
	return ecKey, nil
}
