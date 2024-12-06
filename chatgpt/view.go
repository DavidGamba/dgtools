package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
)

func (m model) View() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.promptView(), m.viewportView(), m.footerView())
}

func (m model) headerView() string {
	var s string
	switch m.mode {
	case textMode:
		s += textModeStyle.Width(m.viewport.Width).Render("Text Mode")
		s += "\n" + textModeStyle.Render(m.helpView())
	case normalMode:
		s += normalModeStyle.Width(m.viewport.Width).Render("Normal Mode")
		s += "\n" + normalModeStyle.Render(m.helpView())
	}
	if m.waiting {
		s += "\n" + m.spinner.View() + " " + m.stopwatch.View()
	} else {
		if len(m.messages) > 0 {
			s += "\n" + " " + m.stopwatch.View()
		} else {
			s += "\n"
		}
	}
	return s
}

func (m model) promptView() string {
	return m.textarea.View()
}

func (m model) viewportView() string {
	return m.viewport.View()
}

func (m model) footerView() string {
	var s string
	switch m.mode {
	case textMode:
		s += textModeStyle.Width(m.viewport.Width).Render("")
	case normalMode:
		s += normalModeStyle.Width(m.viewport.Width).Render("")
	}
	return s
}

func (m model) helpView() string {
	switch m.mode {
	case textMode:
		return m.help.ShortHelpView([]key.Binding{
			m.keymap.esc,
			m.keymap.submit,
		})
	case normalMode:
		return m.help.ShortHelpView([]key.Binding{
			m.keymap.insert,
			m.keymap.showRaw,
			m.keymap.reset,
			m.keymap.quit,
		})
	}
	return ""
}
