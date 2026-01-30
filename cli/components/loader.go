// Package components provides reusable UI components.
package components

import (
	"time"

	"github.com/charmbracelet/bubbletea"
)

// TickMsg is sent for animation ticks.
type TickMsg struct{}

// Loader handles loading animation with a spinner.
type Loader struct {
	frames   []string
	index    int
	active   bool
	deadline time.Time
	text     string
}

// NewLoader creates a new loader with default frames.
func NewLoader() Loader {
	return Loader{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// NewVerifyLoader creates a loader with shorter frames for verification.
func NewVerifyLoader() Loader {
	return Loader{
		frames: []string{"⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start starts the loader with the given text and duration.
func (l Loader) Start(text string, duration time.Duration) (Loader, tea.Cmd) {
	l.active = true
	l.text = text
	l.deadline = time.Now().Add(duration)
	return l, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// Update handles the tick message and advances animation.
func (l Loader) Update() (Loader, tea.Cmd) {
	if !l.active {
		return l, nil
	}

	if time.Now().After(l.deadline) {
		l.active = false
		l.index = 0
		return l, nil
	}

	l.index = (l.index + 1) % len(l.frames)
	return l, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// IsActive returns true if loader is currently animating.
func (l Loader) IsActive() bool {
	return l.active
}

// Frame returns the current animation frame.
func (l Loader) Frame() string {
	if len(l.frames) == 0 {
		return ""
	}
	return l.frames[l.index]
}

// Text returns the loader text.
func (l Loader) Text() string {
	return l.text
}
