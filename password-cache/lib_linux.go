// Package passwordcache provides a utility to save secrets into the linux keyring
package passwordcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"

	"github.com/jsipprell/keyctl"
	"golang.org/x/crypto/ssh/terminal"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

// GetSecret - Gets a secret from the User Session Keyring.
// If the key doesn't exist, it asks the user to enter the password value.
func GetSecret(name, msg string) ([]byte, error) {
	if msg == "" {
		msg = fmt.Sprintf("Enter password for '%s': ", name)
	}

	// Create session
	keyring, err := keyctl.UserSessionKeyring()
	if err != nil {
		return nil, fmt.Errorf("couldn't create keyring session: %w", err)
	}

	// Retrieve
	key, err := keyring.Search(name)
	if err == nil {
		data, err := key.Get()
		if err != nil {
			return nil, fmt.Errorf("couldn't retrieve key data: %w", err)
		}
		info, _ := key.Info()
		logger.Printf("key: %+v", info)
		return data, nil
	}

	// If not found promt user
	fmt.Print(msg)
	password, err := terminal.ReadPassword(int(syscall.Stdin))
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
	// Create session
	userKeyring, err := keyctl.UserSessionKeyring()
	if err != nil {
		return fmt.Errorf("couldn't create keyring user session: %w", err)
	}

	sessionKeyring, err := keyctl.SessionKeyring()
	if err != nil {
		return fmt.Errorf("couldn't create keyring session: %w", err)
	}

	// Store key
	key, err := sessionKeyring.Add(name, password)
	if err != nil {
		return fmt.Errorf("couldn't store '%s' in session keyring: %w", name, err)
	}
	key.ExpireAfter(timeoutSeconds)

	perm := keyctl.PermUserAll | keyctl.PermProcessAll
	if err := keyctl.SetPerm(key, perm); err != nil {
		return fmt.Errorf("couldn't update permissions for '%s' in session keyring: %w", name, err)
	}

	err = keyctl.Link(userKeyring, key)
	if err != nil {
		return fmt.Errorf("couldn't link '%s' to user keyring: %w", name, err)
	}

	err = keyctl.Unlink(sessionKeyring, key)
	if err != nil {
		return fmt.Errorf("couldn't unlink '%s' from session keyring: %w", name, err)
	}

	info, _ := key.Info()
	logger.Printf("key: %+v", info)
	return nil
}

// GetAndCacheSecret - Gets a secret from the User Session Keyring.
// If the key doesn't exist, it asks the user to enter the password value.
// It also saves the secret to the User Session Keyring.
// It will cache the secret for a given number of seconds.
//
// To invalidate a secret, save it with a 1 second timeout.
// Every read will refresh the cache timeout.
func GetAndCacheSecret(name, msg string, timeoutSeconds uint) ([]byte, error) {
	data, err := GetSecret(name, msg)
	if err != nil {
		return data, err
	}

	err = CacheSecret(name, data, timeoutSeconds)
	return data, err
}
