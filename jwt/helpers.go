package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func buildClaims(subject, issuer, expiresInStr string, claims map[string]string) (jwt.MapClaims, error) {
	expiresIn, err := time.ParseDuration(expiresInStr)
	if err != nil {
		return nil, fmt.Errorf("invalid expires-in duration %q: %w", expiresInStr, err)
	}

	now := time.Now()
	mapClaims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(expiresIn).Unix(),
	}
	if subject != "" {
		mapClaims["sub"] = subject
	}
	if issuer != "" {
		mapClaims["iss"] = issuer
	}
	for key, value := range claims {
		mapClaims[key] = value
	}
	return mapClaims, nil
}

func loadRSAPrivateKey(pemFile string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem file: no PEM block found")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an RSA key")
	}
	return rsaKey, nil
}
