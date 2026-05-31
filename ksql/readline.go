package main

import (
	"context"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/hymkor/go-multiline-ny"
	"github.com/hymkor/go-multiline-ny/completion"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-ny/simplehistory"
)

type OSClipboard struct{}

func (OSClipboard) Read() (string, error) {
	return clipboard.ReadAll()
}

func (OSClipboard) Write(s string) error {
	return clipboard.WriteAll(s)
}

var (
	commands = []string{"select", "insert", "delete", "update"}
	tables   = []string{"dept", "emp", "bonus", "salgrade"}
	columns  = []string{"deptno", "dname", "loc", "empno", "ename", "job", "mgr", "hiredate", "sal", "comm", "grade", "losal", "hisal"}
)

func getCompletionCandidates(fields []string) (forCompletion []string, forListing []string) {
	candidates := commands
	for _, word := range fields {
		if strings.EqualFold(word, "from") {
			candidates = append([]string{"where"}, tables...)
		} else if strings.EqualFold(word, "set") {
			candidates = append([]string{"where"}, columns...)
		} else if strings.EqualFold(word, "update") {
			candidates = append([]string{"set"}, tables...)
		} else if strings.EqualFold(word, "delete") {
			candidates = []string{"from"}
		} else if strings.EqualFold(word, "select") {
			candidates = append([]string{"from"}, columns...)
		} else if strings.EqualFold(word, "where") {
			candidates = append([]string{"and", "or"}, columns...)
		}
	}
	return candidates, candidates
}

func interactive(ctx context.Context) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		fmt.Println("C-m or Enter      : Insert a linefeed")
		fmt.Println("C-p or UP         : Move to the previous line.")
		fmt.Println("C-n or DOWN       : Move to the next line")
		fmt.Println("C-j or Meta+Enter : Submit")
		fmt.Println("C-c               : Abort.")
		fmt.Println("C-D with no chars : Quit.")
		fmt.Println("C-UP   or Meta-P  : Move to the previous history entry")
		fmt.Println("C-DOWN or Meta-N  : Move to the next history entry")

		var ed multiline.Editor
		ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
			return fmt.Fprintf(w, "[%d] ", lnum+1)
		})

		ed.SubmitOnEnterWhen(func(lines []string, _ int) bool {
			return strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), ";")
		})

		ed.SetPredictColor(readline.PredictColorBlueItalic)

		ed.Highlight = []readline.Highlight{
			// Words
			{Pattern: regexp.MustCompile(`(?i)(SELECT|INSERT|FROM|WHERE|AS|GROUP BY|ORDER BY|LIMIT)`), Sequence: "\x1B[36;49;1m"},
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
		ed.ResetColor = "\x1B[0m"
		ed.DefaultColor = "\x1B[37;49;1m"

		// enable history (optional)
		history := simplehistory.New()
		ed.SetHistory(history)
		ed.SetHistoryCycling(true)

		// enable completion (optional)
		ed.BindKey(keys.CtrlI, &completion.CmdCompletionOrList{
			// Characters listed here are excluded from completion.
			Delimiter: "&|><;",
			// Enclose candidates with these characters when they contain spaces
			Enclosure: `"'`,
			// String to append when only one candidate remains
			Postfix: " ",
			// Function for listing candidates
			Candidates: getCompletionCandidates,
		})

		// Show newline mark (experimental)
		ed.OnAfterRender = func(w io.Writer, availWidth int) {
			if availWidth >= 2 {
				io.WriteString(w, "\x1B[33;22m↓\x1B[39m")
			}
		}

		for {
			lines, err := ed.Read(ctx)
			if err != nil {
				yield(lines, fmt.Errorf("failed to read: %w", err))
				return
			}
			L := strings.Join(lines, "\n")
			history.Add(L)
			if !yield(lines, nil) {
				return
			}
		}
	}
}
