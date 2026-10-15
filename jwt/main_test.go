package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------

// captureStdout redirects os.Stdout for the duration of f and returns
// everything written to it.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	outC := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	_ = w.Close()
	os.Stdout = old
	return <-outC
}

// captureStderr redirects os.Stderr for the duration of f and returns
// everything written to it.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	outC := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	_ = w.Close()
	os.Stderr = old
	return <-outC
}

func writePEMFile(t *testing.T, dir, name, blockType string, der []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	block := &pem.Block{Type: blockType, Bytes: der}
	data := pem.EncodeToMemory(block)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write pem file %s: %v", path, err)
	}
	return path
}

type rsaKeyPair struct {
	priv *rsa.PrivateKey
}

func genRSAKeyPair(t *testing.T) rsaKeyPair {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	return rsaKeyPair{priv: priv}
}

type ecKeyPair struct {
	priv *ecdsa.PrivateKey
}

func genECKeyPair(t *testing.T) ecKeyPair {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ec key: %v", err)
	}
	return ecKeyPair{priv: priv}
}

func selfSignedCert(t *testing.T, priv interface{}, pub interface{}, cn string) []byte {
	t.Helper()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	return der
}

// ---------------------------------------------------------------------
// buildClaims
// ---------------------------------------------------------------------

func TestBuildClaims(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		claims, err := buildClaims("", "", "1h", map[string]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := claims["sub"]; ok {
			t.Errorf("expected no sub claim, got %v", claims["sub"])
		}
		if _, ok := claims["iss"]; ok {
			t.Errorf("expected no iss claim, got %v", claims["iss"])
		}
		iat, ok := claims["iat"].(int64)
		if !ok {
			t.Fatalf("expected iat to be int64, got %T", claims["iat"])
		}
		exp, ok := claims["exp"].(int64)
		if !ok {
			t.Fatalf("expected exp to be int64, got %T", claims["exp"])
		}
		if diff := exp - iat; diff != 3600 {
			t.Errorf("expected exp-iat to be 3600, got %d", diff)
		}
	})

	t.Run("full", func(t *testing.T) {
		claims, err := buildClaims("subj1", "iss1", "30m", map[string]string{"role": "admin", "team": "eng"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims["sub"] != "subj1" {
			t.Errorf("expected sub=subj1, got %v", claims["sub"])
		}
		if claims["iss"] != "iss1" {
			t.Errorf("expected iss=iss1, got %v", claims["iss"])
		}
		if claims["role"] != "admin" {
			t.Errorf("expected role=admin, got %v", claims["role"])
		}
		if claims["team"] != "eng" {
			t.Errorf("expected team=eng, got %v", claims["team"])
		}
		iat, _ := claims["iat"].(int64)
		exp, _ := claims["exp"].(int64)
		if diff := exp - iat; diff != 1800 {
			t.Errorf("expected exp-iat to be 1800, got %d", diff)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		_, err := buildClaims("", "", "not-a-duration", map[string]string{})
		if err == nil {
			t.Fatal("expected error for invalid duration, got nil")
		}
	})
}

// ---------------------------------------------------------------------
// loadRSAPrivateKey / loadRSAPublicKey
// ---------------------------------------------------------------------

func TestLoadRSAPrivateKey(t *testing.T) {
	dir := t.TempDir()
	kp := genRSAKeyPair(t)

	t.Run("PKCS1", func(t *testing.T) {
		path := writePEMFile(t, dir, "pkcs1.pem", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(kp.priv))
		key, err := loadRSAPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.Cmp(kp.priv.N) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("PKCS8", func(t *testing.T) {
		der, err := x509.MarshalPKCS8PrivateKey(kp.priv)
		if err != nil {
			t.Fatalf("failed to marshal pkcs8: %v", err)
		}
		path := writePEMFile(t, dir, "pkcs8.pem", "PRIVATE KEY", der)
		key, err := loadRSAPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.Cmp(kp.priv.N) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadRSAPrivateKey(filepath.Join(dir, "does-not-exist.pem"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := filepath.Join(dir, "invalid.pem")
		if err := os.WriteFile(path, []byte("not a pem file"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		_, err := loadRSAPrivateKey(path)
		if err == nil {
			t.Fatal("expected error for invalid pem, got nil")
		}
	})

	t.Run("wrong key type", func(t *testing.T) {
		ecKp := genECKeyPair(t)
		der, err := x509.MarshalPKCS8PrivateKey(ecKp.priv)
		if err != nil {
			t.Fatalf("failed to marshal ec pkcs8: %v", err)
		}
		path := writePEMFile(t, dir, "ec-as-rsa.pem", "PRIVATE KEY", der)
		_, err = loadRSAPrivateKey(path)
		if err == nil {
			t.Fatal("expected error when loading an EC key as RSA, got nil")
		}
	})
}

func TestLoadRSAPublicKey(t *testing.T) {
	dir := t.TempDir()
	kp := genRSAKeyPair(t)

	t.Run("PKIX", func(t *testing.T) {
		der, err := x509.MarshalPKIXPublicKey(&kp.priv.PublicKey)
		if err != nil {
			t.Fatalf("failed to marshal pkix: %v", err)
		}
		path := writePEMFile(t, dir, "pkix.pem", "PUBLIC KEY", der)
		key, err := loadRSAPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.Cmp(kp.priv.N) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("PKCS1", func(t *testing.T) {
		der := x509.MarshalPKCS1PublicKey(&kp.priv.PublicKey)
		path := writePEMFile(t, dir, "pkcs1-pub.pem", "RSA PUBLIC KEY", der)
		key, err := loadRSAPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.Cmp(kp.priv.N) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("certificate", func(t *testing.T) {
		der := selfSignedCert(t, kp.priv, &kp.priv.PublicKey, "rsa-test")
		path := writePEMFile(t, dir, "cert.pem", "CERTIFICATE", der)
		key, err := loadRSAPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.Cmp(kp.priv.N) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("certificate with EC key errors", func(t *testing.T) {
		ecKp := genECKeyPair(t)
		der := selfSignedCert(t, ecKp.priv, &ecKp.priv.PublicKey, "ec-test")
		path := writePEMFile(t, dir, "ec-cert.pem", "CERTIFICATE", der)
		_, err := loadRSAPublicKey(path)
		if err == nil {
			t.Fatal("expected error when loading an EC certificate as RSA, got nil")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadRSAPublicKey(filepath.Join(dir, "does-not-exist.pem"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := filepath.Join(dir, "invalid-pub.pem")
		if err := os.WriteFile(path, []byte("not a pem file"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		_, err := loadRSAPublicKey(path)
		if err == nil {
			t.Fatal("expected error for invalid pem, got nil")
		}
	})
}

// ---------------------------------------------------------------------
// loadECPrivateKey / loadECPublicKey
// ---------------------------------------------------------------------

func TestLoadECPrivateKey(t *testing.T) {
	dir := t.TempDir()
	kp := genECKeyPair(t)

	t.Run("SEC1", func(t *testing.T) {
		der, err := x509.MarshalECPrivateKey(kp.priv)
		if err != nil {
			t.Fatalf("failed to marshal sec1: %v", err)
		}
		path := writePEMFile(t, dir, "sec1.pem", "EC PRIVATE KEY", der)
		key, err := loadECPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.X.Cmp(kp.priv.X) != 0 || key.Y.Cmp(kp.priv.Y) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("PKCS8", func(t *testing.T) {
		der, err := x509.MarshalPKCS8PrivateKey(kp.priv)
		if err != nil {
			t.Fatalf("failed to marshal pkcs8: %v", err)
		}
		path := writePEMFile(t, dir, "pkcs8.pem", "PRIVATE KEY", der)
		key, err := loadECPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.X.Cmp(kp.priv.X) != 0 || key.Y.Cmp(kp.priv.Y) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadECPrivateKey(filepath.Join(dir, "does-not-exist.pem"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := filepath.Join(dir, "invalid.pem")
		if err := os.WriteFile(path, []byte("not a pem file"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		_, err := loadECPrivateKey(path)
		if err == nil {
			t.Fatal("expected error for invalid pem, got nil")
		}
	})

	t.Run("wrong key type", func(t *testing.T) {
		rsaKp := genRSAKeyPair(t)
		der, err := x509.MarshalPKCS8PrivateKey(rsaKp.priv)
		if err != nil {
			t.Fatalf("failed to marshal rsa pkcs8: %v", err)
		}
		path := writePEMFile(t, dir, "rsa-as-ec.pem", "PRIVATE KEY", der)
		_, err = loadECPrivateKey(path)
		if err == nil {
			t.Fatal("expected error when loading an RSA key as EC, got nil")
		}
	})
}

func TestLoadECPublicKey(t *testing.T) {
	dir := t.TempDir()
	kp := genECKeyPair(t)

	t.Run("PKIX", func(t *testing.T) {
		der, err := x509.MarshalPKIXPublicKey(&kp.priv.PublicKey)
		if err != nil {
			t.Fatalf("failed to marshal pkix: %v", err)
		}
		path := writePEMFile(t, dir, "pkix.pem", "PUBLIC KEY", der)
		key, err := loadECPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.X.Cmp(kp.priv.X) != 0 || key.Y.Cmp(kp.priv.Y) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("certificate", func(t *testing.T) {
		der := selfSignedCert(t, kp.priv, &kp.priv.PublicKey, "ec-test")
		path := writePEMFile(t, dir, "cert.pem", "CERTIFICATE", der)
		key, err := loadECPublicKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.X.Cmp(kp.priv.X) != 0 || key.Y.Cmp(kp.priv.Y) != 0 {
			t.Errorf("loaded key does not match original")
		}
	})

	t.Run("certificate with RSA key errors", func(t *testing.T) {
		rsaKp := genRSAKeyPair(t)
		der := selfSignedCert(t, rsaKp.priv, &rsaKp.priv.PublicKey, "rsa-test")
		path := writePEMFile(t, dir, "rsa-cert.pem", "CERTIFICATE", der)
		_, err := loadECPublicKey(path)
		if err == nil {
			t.Fatal("expected error when loading an RSA certificate as EC, got nil")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadECPublicKey(filepath.Join(dir, "does-not-exist.pem"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := filepath.Join(dir, "invalid-pub.pem")
		if err := os.WriteFile(path, []byte("not a pem file"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		_, err := loadECPublicKey(path)
		if err == nil {
			t.Fatal("expected error for invalid pem, got nil")
		}
	})
}

// ---------------------------------------------------------------------
// program() end-to-end tests
// ---------------------------------------------------------------------

func extractToken(t *testing.T, stdout string) string {
	t.Helper()
	token := strings.TrimSpace(stdout)
	if token == "" {
		t.Fatalf("expected a token in stdout, got empty output")
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("expected output to look like a JWT, got: %q", token)
	}
	return token
}

func TestProgramHMACRoundTrip(t *testing.T) {
	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "mysecret", "-s", "user1", "-i", "iss1", "-c", "role=admin"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	verifyStdout := captureStdout(t, func() {
		verifyRc = program([]string{"jwt", "verify", "-k", "mysecret", token})
	})
	if verifyRc != 0 {
		t.Fatalf("expected exit code 0, got %d", verifyRc)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal([]byte(verifyStdout), &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v\noutput: %s", err, verifyStdout)
	}
	if claims["sub"] != "user1" {
		t.Errorf("expected sub=user1, got %v", claims["sub"])
	}
	if claims["iss"] != "iss1" {
		t.Errorf("expected iss=iss1, got %v", claims["iss"])
	}
	if claims["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", claims["role"])
	}
}

func TestProgramHMACMissingSecret(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		rc = program([]string{"jwt", "hmac", "-s", "user1"})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code for missing required --secret")
	}
}

func TestProgramHMACInvalidExpiresIn(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "s", "--expires-in=not-a-duration"})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code for invalid --expires-in")
	}
}

func TestProgramVerifyWrongSecret(t *testing.T) {
	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "mysecret", "-s", "user1"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			verifyRc = program([]string{"jwt", "verify", "-k", "wrongsecret", token})
		})
	})
	if verifyRc == 0 {
		t.Fatal("expected non-zero exit code when verifying with the wrong secret")
	}
}

func TestProgramVerifyNoKeyProvided(t *testing.T) {
	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "mysecret", "-s", "user1"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			verifyRc = program([]string{"jwt", "verify", token})
		})
	})
	if verifyRc == 0 {
		t.Fatal("expected non-zero exit code when verifying without a secret")
	}
}

func TestProgramVerifyExpiredToken(t *testing.T) {
	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "mysecret", "--expires-in=-1h"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			verifyRc = program([]string{"jwt", "verify", "-k", "mysecret", token})
		})
	})
	if verifyRc == 0 {
		t.Fatal("expected non-zero exit code when verifying an expired token")
	}
}

func TestProgramVerifyMalformedToken(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			rc = program([]string{"jwt", "verify", "-k", "s", "not.a.validtoken"})
		})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code for malformed token")
	}
}

func TestProgramVerifyMissingTokenArg(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		rc = program([]string{"jwt", "verify", "-k", "s"})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code when the token argument is missing")
	}
}

func TestProgramRSARoundTrip(t *testing.T) {
	dir := t.TempDir()
	kp := genRSAKeyPair(t)
	privPath := writePEMFile(t, dir, "rsa_priv.pem", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(kp.priv))
	pubDER, err := x509.MarshalPKIXPublicKey(&kp.priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPath := writePEMFile(t, dir, "rsa_pub.pem", "PUBLIC KEY", pubDER)

	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "rsa", "-p", privPath, "-s", "user1", "-c", "role=admin"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	verifyStdout := captureStdout(t, func() {
		verifyRc = program([]string{"jwt", "verify", "-p", pubPath, token})
	})
	if verifyRc != 0 {
		t.Fatalf("expected exit code 0, got %d", verifyRc)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal([]byte(verifyStdout), &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v\noutput: %s", err, verifyStdout)
	}
	if claims["sub"] != "user1" {
		t.Errorf("expected sub=user1, got %v", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", claims["role"])
	}
}

func TestProgramRSAMissingPemFile(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		rc = program([]string{"jwt", "rsa", "-s", "user1"})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code for missing required --pem-file")
	}
}

func TestProgramRSAVerifyWrongKey(t *testing.T) {
	dir := t.TempDir()
	kp1 := genRSAKeyPair(t)
	kp2 := genRSAKeyPair(t)

	privPath := writePEMFile(t, dir, "priv1.pem", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(kp1.priv))
	pub2DER, err := x509.MarshalPKIXPublicKey(&kp2.priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pub2Path := writePEMFile(t, dir, "pub2.pem", "PUBLIC KEY", pub2DER)

	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "rsa", "-p", privPath, "-s", "user1"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			verifyRc = program([]string{"jwt", "verify", "-p", pub2Path, token})
		})
	})
	if verifyRc == 0 {
		t.Fatal("expected non-zero exit code when verifying with the wrong public key")
	}
}

func TestProgramECRoundTrip(t *testing.T) {
	dir := t.TempDir()
	kp := genECKeyPair(t)
	privDER, err := x509.MarshalECPrivateKey(kp.priv)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	privPath := writePEMFile(t, dir, "ec_priv.pem", "EC PRIVATE KEY", privDER)
	pubDER, err := x509.MarshalPKIXPublicKey(&kp.priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPath := writePEMFile(t, dir, "ec_pub.pem", "PUBLIC KEY", pubDER)

	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "ec", "-p", privPath, "-s", "user1", "-c", "role=admin"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	var verifyRc int
	verifyStdout := captureStdout(t, func() {
		verifyRc = program([]string{"jwt", "verify", "-p", pubPath, token})
	})
	if verifyRc != 0 {
		t.Fatalf("expected exit code 0, got %d", verifyRc)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal([]byte(verifyStdout), &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v\noutput: %s", err, verifyStdout)
	}
	if claims["sub"] != "user1" {
		t.Errorf("expected sub=user1, got %v", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", claims["role"])
	}
}

func TestProgramECMissingPemFile(t *testing.T) {
	var rc int
	_ = captureStderr(t, func() {
		rc = program([]string{"jwt", "ec", "-s", "user1"})
	})
	if rc == 0 {
		t.Fatal("expected non-zero exit code for missing required --pem-file")
	}
}

func TestProgramCrossAlgorithmVerifyFails(t *testing.T) {
	dir := t.TempDir()
	ecKp := genECKeyPair(t)
	pubDER, err := x509.MarshalPKIXPublicKey(&ecKp.priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	ecPubPath := writePEMFile(t, dir, "ec_pub.pem", "PUBLIC KEY", pubDER)

	var rc int
	stdout := captureStdout(t, func() {
		rc = program([]string{"jwt", "hmac", "-k", "mysecret", "-s", "user1"})
	})
	if rc != 0 {
		t.Fatalf("expected exit code 0, got %d", rc)
	}
	token := extractToken(t, stdout)

	// Verifying an HMAC-signed token with an EC public key should fail:
	// the keyFunc detects the token was signed with an unrelated method.
	var verifyRc int
	_ = captureStderr(t, func() {
		_ = captureStdout(t, func() {
			verifyRc = program([]string{"jwt", "verify", "-p", ecPubPath, token})
		})
	})
	if verifyRc == 0 {
		t.Fatal("expected non-zero exit code when verifying an HMAC token with an EC key")
	}
}

func TestProgramHelp(t *testing.T) {
	// Note: go-getoptions writes help output using a writer captured at
	// package init time, so it cannot be intercepted here by swapping
	// os.Stdout/os.Stderr. We only assert on the exit code.
	for _, cmd := range []string{"hmac", "rsa", "ec", "verify"} {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			var rc int
			_ = captureStderr(t, func() {
				_ = captureStdout(t, func() {
					rc = program([]string{"jwt", "help", cmd})
				})
			})
			if rc == 0 {
				t.Errorf("expected help command to return exit code 1 (getoptions convention), got 0")
			}
		})
	}
}
