// Package common provides shared utilities for the agent-telegram tool.
//revive:disable:var-naming
package common

import "github.com/charmbracelet/lipgloss"

var (
	// TelegramBlue is the Telegram signature blue color.
	TelegramBlue = lipgloss.Color("#0088cc") // Telegram signature blue
	// NormalColor is medium gray for secondary text.
	NormalColor = lipgloss.Color("#6c6c6c") // Medium gray for secondary text
	// ErrorColor is red for errors.
	ErrorColor = lipgloss.Color("#ff5555") // Red for errors

	// TitleStyle is the style for titles.
	TitleStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue).
			Bold(true)

	// LabelStyle is the style for labels.
	LabelStyle = lipgloss.NewStyle().
			Foreground(NormalColor)

	// HelpStyle is the style for help text.
	HelpStyle = lipgloss.NewStyle().
			Foreground(NormalColor)

	// ErrorStyle is the style for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	// InputStyle is the style for input text.
	InputStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue)

	// InputPlaceholderStyle is the style for placeholder text.
	InputPlaceholderStyle = lipgloss.NewStyle().
			Foreground(NormalColor).
			Faint(true)

	// InputFocusedStyle is the style for focused input.
	InputFocusedStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue).
			Bold(true)
)
