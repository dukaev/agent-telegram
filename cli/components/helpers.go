// Package components provides reusable UI components.
package components

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"agent-telegram/pkg/common"
)

// KeyEnter is the enter key string.
const KeyEnter = "enter"

// ContainerStyle adds left padding to the entire view.
var ContainerStyle = lipgloss.NewStyle().PaddingLeft(2)

// getCredentialsSubtitle returns a subtitle indicating whether custom or default API credentials are used.
func getCredentialsSubtitle() string {
	appID := os.Getenv("TELEGRAM_APP_ID")
	appHash := os.Getenv("TELEGRAM_APP_HASH")
	if appID != "" && appHash != "" {
		return common.HelpStyle.Render("Using custom API credentials")
	}
	return common.HelpStyle.Render("Using default API credentials")
}

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
func RenderLoaderView(label, value string) string {
	lines := []string{
		common.TitleStyle.Render("Telegram Login"),
		getCredentialsSubtitle(),
		"",
		common.LabelStyle.Render(label),
		value,
	}
	return ContainerStyle.Render(strings.Join(lines, "\n"))
}

// RenderInputView renders a standard input view with optional help text.
func RenderInputView(inputLine string, showHelp bool) string {
	return RenderInputViewWithError(inputLine, "", showHelp)
}

// RenderInputViewWithError renders a standard input view with error and optional help text.
func RenderInputViewWithError(inputLine, errMsg string, showHelp bool) string {
	lines := []string{
		common.TitleStyle.Render("Telegram Login"),
		getCredentialsSubtitle(),
		"",
		inputLine,
	}

	if errMsg != "" {
		lines = append(lines, RenderErrorBox(errMsg))
	}

	if showHelp {
		lines = append(lines, "")
		lines = append(lines, common.HelpStyle.Render("enter: submit • q: quit"))
	}

	return ContainerStyle.Render(strings.Join(lines, "\n"))
}

// RenderLabeledInput renders a labeled input field with label above.
func RenderLabeledInput(label, input string) string {
	return common.LabelStyle.Render(label) + "\n" + input
}

// ErrorBoxStyle is the style for error boxes.
var ErrorBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(common.ErrorColor).
	Foreground(common.ErrorColor).
	Padding(0, 1)

// RenderErrorBox renders an error message in a styled box.
func RenderErrorBox(errMsg string) string {
	if errMsg == "" {
		return ""
	}
	return ErrorBoxStyle.Render("✗ " + errMsg)
}
