// Package components provides reusable UI components.
package components

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"agent-telegram/pkg/common"
)

// HandleQuitKeys returns tea.Quit if msg is a quit key.
// Returns ok=true if key was handled, ok=false otherwise.
func HandleQuitKeys(msg tea.Msg) (tea.Cmd, bool) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return tea.Quit, true
		}
	}
	return nil, false
}

// RenderLoaderView renders the loading state with spinner.
func RenderLoaderView(label, value, frame string) string {
	labelWithSpinner := common.LabelStyle.Render(label) + " " +
		lipgloss.NewStyle().Foreground(common.TelegramBlue).Render(frame)

	lines := []string{
		common.TitleStyle.Render(" Telegram Login"),
		"",
		labelWithSpinner,
		value,
	}
	return strings.Join(lines, "\n")
}

// RenderInputView renders a standard input view with optional help text.
func RenderInputView(inputLine string, showHelp bool) string {
	lines := []string{
		common.TitleStyle.Render(" Telegram Login"),
		"",
		inputLine,
	}

	if showHelp {
		lines = append(lines, "")
		lines = append(lines, common.HelpStyle.Render("enter: submit • q: quit"))
	}

	return strings.Join(lines, "\n")
}

// RenderLabeledInput renders a labeled input field with label above.
func RenderLabeledInput(label, input string) string {
	return common.LabelStyle.Render(label) + "\n" + input
}
