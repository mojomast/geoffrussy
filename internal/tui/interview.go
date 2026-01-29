package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InterviewModel represents the TUI model for the interview process
// NOTE: This is a stub implementation. Full integration requires:
// - Proper session management with the interview engine
// - State synchronization
// - Error handling
type InterviewModel struct {
	message  string
	quitting bool
}

// NewInterviewModel creates a new interview TUI model
func NewInterviewModel() InterviewModel {
	return InterviewModel{
		message: "Interview TUI - Implementation in progress",
	}
}

// Init initializes the interview model
func (m InterviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m InterviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the interview UI
func (m InterviewModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	return headerStyle.Render("🎤 Geoffrey Interview") + "\n\n" +
		m.message + "\n\n" +
		"Press Q to quit\n"
}

// StartInterview is a stub for starting the interview TUI
func StartInterview() error {
	model := NewInterviewModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running interview: %w", err)
	}
	return nil
}
