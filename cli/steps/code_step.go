// Package steps provides login step implementations.
package steps

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"agent-telegram/cli/components"
	"agent-telegram/pkg/common"
)

// CodeStep represents the verification code input step.
type CodeStep struct {
	codeInput components.MaskedInput
	loader    components.Loader
}

// NewCodeStep creates a new code input step.
func NewCodeStep() CodeStep {
	codeInput := components.NewMaskedInput(5, components.CodeType)
	codeInput.Focus()

	return CodeStep{
		codeInput: codeInput,
		loader:    components.NewVerifyLoader(),
	}
}

// Init initializes the step.
func (m CodeStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m CodeStep) Update(msg tea.Msg) (CodeStep, tea.Cmd) {
	if cmd, ok := components.HandleQuitKeys(msg); ok {
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok && !m.loader.IsActive() {
		switch msg.String() {
		case "enter":
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Start("Verifying code...", 2*time.Second)
			return m, cmd
		}
	}

	if _, ok := msg.(components.TickMsg); ok {
		if m.loader.IsActive() {
			var cmd tea.Cmd
			m.loader, cmd = m.loader.Update()
			if !m.loader.IsActive() {
				return m, m.Submit()
			}
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.codeInput, cmd = m.codeInput.Update(msg)
	return m, cmd
}

// Submit returns the verification code.
func (m CodeStep) Submit() tea.Cmd {
	return func() tea.Msg {
		return CodeSubmitted(m.codeInput.Value())
	}
}

// View renders the code input step.
func (m CodeStep) View() string {
	if m.loader.IsActive() {
		return components.RenderLoaderView(
			m.codeInput.GetLabel(),
			m.codeInput.ValueOnlyColored(common.TelegramBlue),
			m.loader.Frame(),
		)
	}

	inputLine := components.RenderLabeledInput(m.codeInput.GetLabel(), m.codeInput.View())
	return components.RenderInputView(inputLine, true)
}

// GetCode returns the entered code.
func (m CodeStep) GetCode() string {
	return m.codeInput.Value()
}

// CodeSubmitted is sent when code is submitted.
type CodeSubmitted string
