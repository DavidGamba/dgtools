package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", `Operations on patch files.

         Source: https://github.com/DavidGamba/dgtools`)
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))

	split := opt.NewCommand("split", `Split a patch by hunks
NOTE: applying an individual subset might invalidate subsequent hunks`)
	split.HelpSynopsisArgs("<file>...")
	split.SetCommandFn(SplitRun)
	split.Bool("keep", false, opt.Description("Keep source patch file"))
	split.Bool("verbose", false, opt.Description("Print hunks when splitting file"))

	selectCmd := opt.NewCommand("select", `Go over multiple patch files and select which ones to keep
NOTE: patches that are not selected are deleted.`)
	selectCmd.Bool("reverse", false, opt.Description("Reverse patch selection"), opt.Alias("R"))
	selectCmd.SetCommandFn(SelectRun)

	// TODO: Join patch files before applying them otherwise they might affect the original line numbering

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
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

func SplitRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	keep := opt.Value("keep").(bool)
	verbose := opt.Value("verbose").(bool)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}
	for _, file := range args {
		err := SplitPatchFile(ctx, file, verbose, keep)
		if err != nil {
			return fmt.Errorf("failed to split patch file '%s': %w", file, err)
		}
	}

	return nil
}

func SelectRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	reverse := opt.Value("reverse").(bool)

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing <file>\n")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	for _, file := range args {
		err := SelectPatchFile(ctx, file, reverse)
		if err != nil {
			return fmt.Errorf("failed to split patch file '%s': %w", file, err)
		}
	}

	return nil
}

func UserConfirmation(msg string) (bool, error) {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	rsp := scanner.Text()
	if rsp == "N" || rsp == "n" {
		return false, nil
	}
	if rsp == "Y" || rsp == "y" || rsp == "" {
		return true, nil
	}
	return false, fmt.Errorf("invalid input")
}

type PatchHunks struct {
	File   string
	Header string
	Hunks  []string
	Size   int64
}

func GetPatchFileHunks(ctx context.Context, file string) (*PatchHunks, error) {
	ph := &PatchHunks{}
	ph.File = file

	fh, err := os.Open(file)
	if err != nil {
		return ph, fmt.Errorf("failed to open file: %w", err)
	}
	defer fh.Close()

	fi, err := fh.Stat()
	if err != nil {
		return ph, fmt.Errorf("failed to stat file: %w", err)
	}
	ph.Size = fi.Size()
	if ph.Size == 0 {
		return ph, nil
	}

	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	c := 0
	header := ""
	hunks := []string{}
	currentHunk := ""
	for scanner.Scan() {
		c++
		if c <= 3 {
			header += scanner.Text() + "\n"
		} else {
			line := scanner.Text()
			if strings.HasPrefix(line, "@@ ") && strings.HasSuffix(line, " @@") {
				// first hunk
				if currentHunk == "" {
					currentHunk = line + "\n"
					continue
				}
				// new hunk
				hunks = append(hunks, currentHunk)
				currentHunk = line + "\n"
				continue
			}
			// hunk continuation
			currentHunk += line + "\n"
		}
	}
	// save last hunk
	hunks = append(hunks, currentHunk)

	Logger.Printf("Found %d hunks in %s\n", len(hunks), file)
	ph.Header = header
	ph.Hunks = hunks

	return ph, nil
}

func SplitPatchFile(ctx context.Context, file string, verbose, keep bool) error {
	outputFile := strings.TrimSuffix(file, ".patch")

	ph, err := GetPatchFileHunks(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to get hunks for '%s': %w", file, err)
	}

	if len(ph.Hunks) <= 1 {
		Logger.Printf("Nothing to do for %s\n", file)
		return nil
	}
	for i, hunk := range ph.Hunks {
		file := fmt.Sprintf("%s-%02d.patch", outputFile, i+1)
		Logger.Printf("hunk: %s\n", file)
		if verbose {
			fmt.Printf("%s", ph.Header)
			fmt.Printf("%s", hunk)
		}
		err := os.WriteFile(file, []byte(ph.Header+hunk), 0640)
		if err != nil {
			return fmt.Errorf("failed to write to %s: %w", file, err)
		}
	}
	if !keep {
		os.Remove(file)
	}

	return nil
}

func SelectPatchFile(ctx context.Context, file string, reverse bool) error {
	outputFile := strings.TrimSuffix(file, ".patch") + ".fpatch"

	ph, err := GetPatchFileHunks(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to get hunks for '%s': %w", file, err)
	}
	if ph.Size == 0 {
		Logger.Printf("empty patch, deleting file '%s'\n", file)
		os.Remove(file)
		return nil
	}

	oh, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer oh.Close()
	Logger.Printf("output file: %s\n", oh.Name())

	fmt.Printf("%s", ph.Header)

	headerAdded := false
	for i, hunk := range ph.Hunks {
		hname := fmt.Sprintf("%s %02d.patch", outputFile, i+1)
		Logger.Printf("hunk: %s\n", hname)
		fmt.Printf("%s", hunk)
		keep, err := UserConfirmation("Keep hunk [Y/n]: ")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}
		if !reverse {
			if keep {
				if !headerAdded {
					Logger.Printf("Adding header to %s\n", oh.Name())
					fmt.Fprintf(oh, "%s", ph.Header)
					headerAdded = true
				}
				Logger.Printf("Adding hunk to %s\n", oh.Name())
				fmt.Fprintf(oh, "%s", hunk)
			}
		} else {
			if !keep {
				if !headerAdded {
					Logger.Printf("Adding header to %s\n", oh.Name())
					fmt.Fprintf(oh, "%s", ph.Header)
					headerAdded = true
				}
				Logger.Printf("Adding hunk to %s\n", oh.Name())
				fmt.Fprintf(oh, "%s", hunk)
			}
		}
	}

	ohi, err := oh.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat: %w", err)
	}
	Logger.Printf("patch file size %d\n", ohi.Size())
	if ohi.Size() == 0 {
		os.Remove(oh.Name())
	}

	return nil
}
