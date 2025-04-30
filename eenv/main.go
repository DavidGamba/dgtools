package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(io.Discard, "", 0)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Print your environment like env but with the keys, passwords and tokens hidden

   Source: https://github.com/DavidGamba/dgtools`)
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("debug", false)
	opt.Float64("entropy-threshold", 4.0, opt.Description("Entropy threshold to hide the value"))
	opt.SetCommandFn(Run)
	opt.HelpSynopsisArg("", "")
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
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
		if errors.Is(err, getoptions.ErrorParsing) {
			fmt.Fprintf(os.Stderr, "\n"+opt.Help())
		}
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	threshold := opt.Value("entropy-threshold").(float64)
	Logger.Printf("Running")

	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		key := strings.ToLower(parts[0])
		value := strings.ToLower(parts[1])

		if slices.Contains(knownSafeVars, parts[0]) {
			fmt.Printf("%s=%s\n", parts[0], parts[1])
			continue
		}

		e := shannonEntropy(value)
		if strings.Contains(key, "key") || strings.Contains(key, "password") || strings.Contains(key, "token") {
			fmt.Printf("%s=%s\n", parts[0], "***")
			Logger.Printf("(%.2f)", e)
			continue
		}
		// known secret prefixes
		if strings.HasPrefix(value, "sk-") || strings.HasPrefix(value, "ghp_") {
			fmt.Printf("%s=%s\n", parts[0], "***")
			Logger.Printf("(%.2f)", e)
			continue
		}
		// avoid printing dev type secrets
		if strings.Contains(key, "key") || strings.Contains(key, "password") || strings.Contains(key, "token") {
			fmt.Printf("%s=%s\n", parts[0], "***")
			Logger.Printf("(%.2f)", e)
			continue
		}

		if e >= threshold {
			fmt.Printf("%s=%s\n", parts[0], "***")
			Logger.Printf("(%.2f)", e)
			continue
		}
		fmt.Printf("%s=%s\n", parts[0], parts[1])
		Logger.Printf("(%.2f)", e)
	}
	return nil
}

var knownSafeVars = []string{
	"PATH",
	"HOME",
	"USER",
	"LANG",

	"TERM",
	"TERMINFO",

	"PWD",
	"OLDPWD",

	"JAVA_HOME",

	"EDITOR",
	"VISUAL",
	"PAGER",
	"MANPATH",
}

// h(x) = -sum(p(x) * log2(p(x)))
// https://en.wikipedia.org/wiki/Entropy_(information_theory)
// log base 2 to determine the number of bits needed to store the information
//
// Approach:
// 1. Get a count of all the symbols in the string: N
// 2. Calculate the frequency of each symbol Sn, which is the number of times the symbol Sn appears divided by the number of symbols. p(Sn) = count(Sn) / N
// 3. Calculate the entropy per symbol: p(Sn)*log2(p(Sn))
// 4. Sum the entropy of all symbols: H(X) = -sum( p(S1)*log2(p(S1)) + p(S2)*log2(p(S2)) + ... + p(Sn)*log2(p(Sn)) )
func shannonEntropy(s string) float64 {
	symbols := make(map[rune]int)
	for _, r := range s {
		symbols[r]++
	}
	n := len(s)
	entropy := 0.0
	for _, count := range symbols {
		p := float64(count) / float64(n)
		entropy += p * math.Log2(p)
	}
	return -entropy
}
