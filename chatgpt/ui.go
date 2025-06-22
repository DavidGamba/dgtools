package main

import (
	"context"
	"strings"
	"time"

	"github.com/DavidGamba/go-getoptions"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func UIRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

// current typing mode
type currentMode uint

const (
	textMode currentMode = iota
	normalMode
)

var (
	width = 200
)

type model struct {
	mode      currentMode
	viewReady bool
	waiting   bool
	showRaw   bool

	messages    []string
	rawMessages []string

	keymap keymap

	// Models
	stopwatch stopwatch.Model
	spinner   spinner.Model
	viewport  viewport.Model
	textarea  textarea.Model
	help      help.Model

	threads *thread

	quitting bool

	err error
}

type keymap struct {
	start   key.Binding
	stop    key.Binding
	reset   key.Binding
	showRaw key.Binding
	tab     key.Binding
	esc     key.Binding
	insert  key.Binding
	submit  key.Binding
	quit    key.Binding

	j key.Binding
	k key.Binding
}

func newModel() model {
	m := model{}
	m.mode = textMode

	m.threads = NewThread()

	// Initialize models
	m.stopwatch = stopwatch.NewWithInterval(time.Millisecond)
	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Globe
	m.spinner.Style = spinnerStyle
	// m.spinner.Spinner = spinner.Points

	ta := textarea.New()
	ta.CharLimit = 0
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "

	ta.SetHeight(3)

	// Remove cursor line styling
	// ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(true)

	m.textarea = ta

	m.keymap = keymap{
		reset: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "reset"),
		),
		showRaw: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "show raw"),
		),
		tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch view"),
		),
		esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Change to normal mode"),
		),
		insert: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "Change to insert mode"),
		),
		submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "submit"),
		),
		k: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k / ↑", "Up"),
		),
		j: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j / ↓", "Down"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q", "quit"),
		),
	}
	m.keymap.stop.SetEnabled(false)
	m.help = help.New()
	m.err = nil
	return m
}

func (m model) Init() tea.Cmd {
	// init the submodels
	return tea.Batch(m.spinner.Tick, textarea.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		promptHeight := lipgloss.Height(m.promptView())
		verticalMarginHeight := headerHeight + footerHeight + promptHeight

		if !m.viewReady {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewReady = true
		}
		m.textarea.SetWidth(msg.Width)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

		return m, nil
	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks but not when in raw mode to allow selecting without a refresh
		if !m.showRaw {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	case tea.KeyMsg:
		switch m.mode {
		case textMode:
			switch {
			case key.Matches(msg, m.keymap.esc):
				m.mode = normalMode
				return m, nil
			case key.Matches(msg, m.keymap.submit):
				v := m.textarea.Value()

				if v == "" {
					// Don't send empty messages.
					return m, nil
				}
				m.messages = append(m.messages, senderStyle.Render("You: ")+v)
				m.rawMessages = append(m.rawMessages, v)

				content := strings.Join(m.messages, "\n")
				str := lipgloss.NewStyle().Width(width).Render(content)
				m.viewport.SetContent(str)

				m.textarea.Reset()
				m.waiting = true
				m.viewport.GotoBottom()

				cmds = append(cmds, m.stopwatch.Reset(), m.stopwatch.Start())
				cmds = append(cmds, m.threads.sendQueryMsg(context.Background(), v))
			default:
				// Send all other keypresses to the textarea.
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		case normalMode:
			switch {
			case key.Matches(msg, m.keymap.quit):
				m.quitting = true
				return m, tea.Quit
			case key.Matches(msg, m.keymap.insert):
				m.mode = textMode
				return m, nil
			case key.Matches(msg, m.keymap.j):
				m.viewport.LineDown(1)
			case key.Matches(msg, m.keymap.k):
				m.viewport.LineUp(1)
			case key.Matches(msg, m.keymap.showRaw):
				m.showRaw = !m.showRaw
				var content string
				if m.showRaw {
					content = strings.Join(m.rawMessages, "\n")
				} else {
					content = strings.Join(m.messages, "\n")
				}
				str := lipgloss.NewStyle().Width(width).Render(content)
				m.viewport.SetContent(str)
			case key.Matches(msg, m.keymap.reset):
				m.textarea.Reset()
				m.viewport.SetContent("")
				m.messages = []string{}
				m.rawMessages = []string{}
			case key.Matches(msg, m.keymap.start, m.keymap.stop):
				m.keymap.stop.SetEnabled(!m.stopwatch.Running())
				m.keymap.start.SetEnabled(m.stopwatch.Running())
				cmds = append(cmds, m.stopwatch.Toggle())
				// cmds = append(cmds, m.stopwatch.Reset())
			}
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case queryMsg:
		if msg.err != nil {
			m.messages = append(m.messages, msg.err.Error())
		} else {
			md, err := glamour.Render(msg.content, "dark")
			if err != nil {
				m.messages = append(m.messages, err.Error())
				m.messages = append(m.messages, msg.content)
				m.rawMessages = append(m.rawMessages, err.Error())
				m.rawMessages = append(m.rawMessages, msg.content)
			}
			m.rawMessages = append(m.rawMessages, msg.content)
			m.messages = append(m.messages, md)
		}
		content := strings.Join(m.messages, "\n")
		str := lipgloss.NewStyle().Width(width).Render(content)
		m.viewport.SetContent(str)

		m.viewport.GotoBottom()
		m.waiting = false

		cmds = append(cmds, m.stopwatch.Stop())
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	m.stopwatch, cmd = m.stopwatch.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
