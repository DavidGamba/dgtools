package main

import (
	"fmt"
	"io"
	"log"
	"os"

	passwordcache "github.com/DavidGamba/dgtools/password-cache"
	"github.com/DavidGamba/go-getoptions"
)

var logger = log.New(io.Discard, "", log.LstdFlags)

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
	opt.HelpSynopsisArg("<key-name>", "Password key name")
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

	data, err := passwordcache.GetAndCacheSecret(keyName, "", uint(timeout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("print") {
		fmt.Printf("%s", data)
	}
	return 0
}
