package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/DavidGamba/go-getoptions"
	"github.com/jsipprell/keyctl"
	"golang.org/x/crypto/ssh/terminal"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	var timeout int
	opt := getoptions.New()
	opt.Self("", `Saves/Retrieves passwords from/to the Linux keyring.
	It will cache the password for a given timeout.
	If a password doesn't exist in the keyring it will prompt the user for one.`)
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.Bool("print", false, opt.Description("Print password to STDOUT"))
	opt.IntVar(&timeout, "timeout", 900, opt.ArgName("seconds"),
		opt.Description("Timeout in seconds, default 15 minutes"))
	opt.HelpSynopsisArgs("<key-name>")
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)
	if len(remaining) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: Missing key\n")
		fmt.Fprintln(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		return 1
	}
	keyName := remaining[0]

	// Retrieve existing password or ask user to add one
	data, err := GetPassword(keyName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}

	err = CachePassword(keyName, string(data), uint(timeout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("print") {
		fmt.Printf("%s", data)
	}
	return 0
}

// GetPassword - Gets a secret from the User Session Keyring.
// If the key doesn't exist, it asks the user to enter the password value.
// It will cache the secret for a given number of seconds.
func GetPassword(name string) ([]byte, error) {
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
	fmt.Printf("Enter password for '%s': ", name)
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	return password, nil
}

// CachePassword - Saves a secret to the User Session Keyring.
// It will cache the secret for a given number of seconds.
//
// To invalidate a password, save it with a 1 second timeout.
func CachePassword(name, password string, timeoutSeconds uint) error {
	// Create session
	keyring, err := keyctl.UserSessionKeyring()
	if err != nil {
		return fmt.Errorf("couldn't create keyring session: %w", err)
	}

	// Store key
	keyring.SetDefaultTimeout(timeoutSeconds)
	key, err := keyring.Add(name, []byte(password))
	if err != nil {
		return fmt.Errorf("couldn't store '%s': %s", name, err)
	}
	info, _ := key.Info()
	logger.Printf("key: %+v", info)
	return nil
}
