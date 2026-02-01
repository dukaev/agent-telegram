// Package ui provides interactive UI components for the agent-telegram tool.
package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"agent-telegram/cli/components"
	"agent-telegram/cli/steps"
	"agent-telegram/internal/auth"
	"agent-telegram/pkg/common"
)

// Step interface represents a login step.
type Step interface {
	Init() tea.Cmd
	View() string
}

// messageHandlerCheck defines a message type check and its handler.
type messageHandlerCheck struct {
	check   func(tea.Msg) bool
	handle  func(LoginModel, tea.Msg) (tea.Model, tea.Cmd)
}

// LoginModel manages the login flow.
type LoginModel struct {
	ctx            context.Context
	authService    *auth.Service
	currentStep    Step
	phone          string
	twoFAHint      string
	quitting       bool
	successMsg     string
	errorMsg       string
	sessionPath    string
	messageHandlers []messageHandlerCheck
}

// NewLoginModel creates a new login model starting with phone step.
func NewLoginModel(ctx context.Context, authService *auth.Service) LoginModel {
	m := LoginModel{
		ctx:         ctx,
		authService: authService,
		currentStep: steps.NewPhoneStep(authService),
		messageHandlers: []messageHandlerCheck{
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(auth.Result); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleAuthResultMsg(msg.(auth.Result)) },
			},
			{
				check: func(msg tea.Msg) bool { _, ok := msg.(steps.PhoneCodeSent); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					return m.handlePhoneCodeSentMsg(msg.(steps.PhoneCodeSent))
				},
			},
			{
				check: func(msg tea.Msg) bool { _, ok := msg.(steps.TwoFARequired); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					return m.handleTwoFARequiredMsg(msg.(steps.TwoFARequired))
				},
			},
			{
				check: func(msg tea.Msg) bool { _, ok := msg.(steps.PhoneSubmitted); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) {
					return m.handlePhoneSubmittedMsg(msg.(steps.PhoneSubmitted))
				},
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.CodeSubmitted); return ok },
				handle: func(m LoginModel, _ tea.Msg) (tea.Model, tea.Cmd) { return m.handleCodeSubmittedMsg() },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.PasswordSubmitted); return ok },
				handle: func(m LoginModel, _ tea.Msg) (tea.Model, tea.Cmd) { return m.handlePasswordSubmittedMsg() },
			},
			{
				check:  func(msg tea.Msg) bool { _, ok := msg.(steps.AuthError); return ok },
				handle: func(m LoginModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.handleAuthErrorMsg(msg.(steps.AuthError)) },
			},
		},
	}
	return m
}

// Init initializes the login model.
func (m LoginModel) Init() tea.Cmd {
	if step := m.currentStep; step != nil {
		return step.Init()
	}
	return nil
}

// Update handles messages and transitions between steps.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	// Handle quit keys at top level
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Find handler for this message type
	for _, handler := range m.messageHandlers {
		if handler.check(msg) {
			return handler.handle(m, msg)
		}
	}

	// Delegate update to current step
	return m.updateCurrentStep(msg)
}

// View renders the current step.
func (m LoginModel) View() string {
	if m.quitting {
		if m.errorMsg != "" {
			return components.ContainerStyle.Render(common.ErrorStyle.Render("✗ Error: " + m.errorMsg))
		}
		if m.successMsg != "" {
			return components.ContainerStyle.Render(common.TitleStyle.Render(m.successMsg))
		}
		return ""
	}

	if m.currentStep != nil {
		return m.currentStep.View()
	}

	return ""
}
