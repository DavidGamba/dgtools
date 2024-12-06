package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	normalModeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FADA7A")).
			Foreground(lipgloss.Color("0")).
			Bold(true)
	textModeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Bold(true)
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	senderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)
