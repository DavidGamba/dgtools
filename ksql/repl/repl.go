package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/hymkor/go-multiline-ny"
	"github.com/hymkor/go-multiline-ny/completion"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
)

type OSClipboard struct{}

func (OSClipboard) Read() (string, error) {
	return clipboard.ReadAll()
}

func (OSClipboard) Write(s string) error {
	return clipboard.WriteAll(s)
}

type HistoryFile struct {
	filename string
}

func (h *HistoryFile) At(n int) string {
	f, err := os.Open(h.filename)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		if i == n {
			return strings.ReplaceAll(scanner.Text(), "⏎", "\n")
		}
		i++
	}
	return ""
}

func (h *HistoryFile) Len() int {
	f, err := os.Open(h.filename)
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func (h *HistoryFile) Add(line string) error {
	f, err := os.OpenFile(h.filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	if err != nil {
		return fmt.Errorf("failed to write to history file: %w", err)
	}
	return nil
}

// NewHistoryFile - creates a new history file in the user's Cache directory.
// dir: the dir within the user's cache directory where to place the history file.
// name: the name of the history file.
func NewHistoryFile(dir, name string) (*HistoryFile, error) {
	cacheDirBase, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache dir: %w", err)
	}
	cacheDir := filepath.Join(cacheDirBase, dir)
	os.MkdirAll(cacheDir, 0755)
	historyFile := filepath.Join(cacheDir, name)
	return &HistoryFile{historyFile}, nil
}

func DefaultHeader() string {
	return `[USAGE]
	C-m or Enter      : Insert a linefeed
	C-p or UP         : Move to the previous line.
	C-n or DOWN       : Move to the next line
	C-j or Meta+Enter : Submit
	C-c               : Abort.
	C-D with no chars : Quit.
	C-UP   or Meta-P  : Move to the previous history entry
	C-DOWN or Meta-N  : Move to the next history entry
`
}

type Repl struct {
	Ed     *multiline.Editor
	Header string
}

func New(history readline.IHistory, completionCandidates func(fieldsBeforeCursor []string) (completionSet []string, listingSet []string)) *Repl {
	var ed multiline.Editor
	ed.ResetColor = "\x1B[0m"
	ed.DefaultColor = "\x1B[37;49;1m"
	ed.SetHistory(history)
	ed.SetHistoryCycling(false)
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "[%d] ", lnum+1)
	})

	// Show newline mark (experimental)
	ed.OnAfterRender = func(w io.Writer, availWidth int) {
		if availWidth >= 2 {
			io.WriteString(w, "\x1B[33;22m⏎\x1B[39m")
		}
	}

	// enable completion (optional)
	ed.BindKey(keys.CtrlI, &completion.CmdCompletionOrList{
		// Characters listed here are excluded from completion.
		Delimiter: "&|><;",
		// Enclose candidates with these characters when they contain spaces
		Enclosure: `"'`,
		// String to append when only one candidate remains
		Postfix: " ",
		// Function for listing candidates
		Candidates: completionCandidates,
	})

	ed.SetPredictColor(readline.PredictColorBlueItalic)

	ed.Highlight = []readline.Highlight{
		// Double quotation
		{Pattern: regexp.MustCompile(`(?m)"([^"\n]*\\")*[^"\n]*$|"([^"\n]*\\")*[^"\n]*"`), Sequence: "\x1B[33;49;1m"},
		// Single quotation
		{Pattern: regexp.MustCompile(`(?m)'([^'\n]*\\')*[^'\n]*$|'([^'\n]*\\')*[^'\n]*'`), Sequence: "\x1B[31;49;1m"},
		// Number literal
		{Pattern: regexp.MustCompile(`[0-9]+`), Sequence: "\x1B[34;49;1m"},
		// Comment
		{Pattern: regexp.MustCompile(`(?s)/\*.*?\*/`), Sequence: "\x1B[30;49;1m"},
		// Multiline string literal
		{Pattern: regexp.MustCompile("(?s)```.*?```"), Sequence: "\x1B[31;49;22m"},
	}
	return &Repl{Header: DefaultHeader(), Ed: &ed}
}

func (r *Repl) SubmitOnEnterWhenEndsOn(s string) {
	r.Ed.SubmitOnEnterWhen(func(lines []string, _ int) bool {
		return strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), s)
	})
}

func Interactive(ctx context.Context, r *Repl) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		fmt.Println(r.Header)

		for {
			lines, err := r.Ed.Read(ctx)
			if err != nil {
				yield(lines, fmt.Errorf("failed to read: %w", err))
				return
			}
			if !yield(lines, nil) {
				return
			}
		}
	}
}
