// Package passwordcache provides a utility to save secrets into the linux keyring
package passwordcache

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/DavidGamba/dgtools/run"
	"golang.org/x/term"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

// GetSecret - Gets a secret from the User Session Keyring.
// If the key doesn't exist, it asks the user to enter the password value.
func GetSecret(name, msg string) ([]byte, error) {
	if msg == "" {
		msg = fmt.Sprintf("Enter password for '%s': ", name)
	}

	out, err := run.CMD("security", "find-generic-password", "-a", os.Getenv("USER"), "-s", name, "-w").STDOutOutput()
	if err == nil {
		secret := strings.TrimSuffix(string(out), "\n")
		return []byte(secret), nil
	}

	// If not found promt user
	fmt.Print(msg)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}
	return password, nil
}

// CacheSecret - Saves a secret to the User Session Keyring.
// It will cache the secret for a given number of seconds.
//
// To invalidate a secret, save it with a 1 second timeout.
func CacheSecret(name string, password []byte, timeoutSeconds uint) error {
	if timeoutSeconds <= 1 {
		return run.CMD("security", "delete-generic-password", "-a", os.Getenv("USER"), "-s", name).Run(io.Discard)
	}
	err := run.CMD("security", "find-generic-password", "-a", os.Getenv("USER"), "-s", name, "-w").Run(io.Discard)
	if err != nil {
		return run.CMD("security", "add-generic-password", "-a", os.Getenv("USER"), "-s", name, "-w", string(password)).Run(io.Discard)
	}
	return nil
}

// GetAndCacheSecret - Gets a secret from the User Session Keyring.
// If the key doesn't exist, it asks the user to enter the password value.
// It also saves the secret to the User Session Keyring.
// It will cache the secret for a given number of seconds.
//
// To invalidate a secret, save it with a 0 second timeout.
// Every read will refresh the cache timeout.
func GetAndCacheSecret(name, msg string, timeoutSeconds uint) ([]byte, error) {
	data, err := GetSecret(name, msg)
	if err != nil {
		return data, err
	}

	err = CacheSecret(name, data, timeoutSeconds)
	return data, err
}
