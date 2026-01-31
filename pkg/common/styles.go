// Package common provides shared utilities for the agent-telegram tool.
package common

import "github.com/charmbracelet/lipgloss"

var (
	// Telegram brand colors
	// Primary: Telegram Blue (#0088cc / #2AABEE)
	// Neutral: Gray tones
	TelegramBlue = lipgloss.Color("#0088cc") // Telegram signature blue
	NormalColor  = lipgloss.Color("#6c6c6c") // Medium gray for secondary text
	ErrorColor   = lipgloss.Color("#ff5555") // Red for errors

	// Telegram brand styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(NormalColor)

	HelpStyle = lipgloss.NewStyle().
			Foreground(NormalColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue)

	InputPlaceholderStyle = lipgloss.NewStyle().
			Foreground(NormalColor).
			Faint(true)

	InputFocusedStyle = lipgloss.NewStyle().
			Foreground(TelegramBlue).
			Bold(true)
)
