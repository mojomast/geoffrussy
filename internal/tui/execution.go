package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExecutionModel represents the TUI model for task execution
type ExecutionModel struct {
	currentPhase   string
	currentTask    string
	output         []string
	status         string
	progress       float64
	paused         bool
	err            error
	quitting       bool
	lastUpdate     time.Time
	maxOutputLines int
}

// NewExecutionModel creates a new execution TUI model
func NewExecutionModel() ExecutionModel {
	return ExecutionModel{
		output:         []string{},
		status:         "initializing",
		progress:       0.0,
		maxOutputLines: 20,
		lastUpdate:     time.Now(),
	}
}

// OutputMsg represents a new output line
type OutputMsg struct {
	Line string
}

// StatusMsg represents a status update
type StatusMsg struct {
	Status string
}

// ProgressMsg represents a progress update
type ProgressMsg struct {
	Progress float64
}

// Init initializes the execution model
func (m ExecutionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+p", "p":
			// Pause execution
			m.paused = !m.paused
			if m.paused {
				m.status = "paused"
			} else {
				m.status = "running"
			}
			return m, nil

		case "s":
			// Skip current task
			m.status = "skipping"
			return m, nil
		}

	case OutputMsg:
		// Add new output line
		m.output = append(m.output, msg.Line)
		// Keep only last N lines
		if len(m.output) > m.maxOutputLines {
			m.output = m.output[len(m.output)-m.maxOutputLines:]
		}
		m.lastUpdate = time.Now()
		return m, nil

	case StatusMsg:
		m.status = msg.Status
		return m, nil

	case ProgressMsg:
		m.progress = msg.Progress
		return m, nil
	}

	return m, nil
}

// View renders the execution UI
func (m ExecutionModel) View() string {
	if m.quitting {
		return "Execution stopped.\n"
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("🚀 Geoffrey Execution"))
	b.WriteString("\n\n")

	// Current phase and task
	phaseStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	if m.currentPhase != "" {
		b.WriteString(phaseStyle.Render(fmt.Sprintf("Phase: %s", m.currentPhase)))
		b.WriteString("\n")
	}

	if m.currentTask != "" {
		taskStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))
		b.WriteString(taskStyle.Render(fmt.Sprintf("Task: %s", m.currentTask)))
		b.WriteString("\n")
	}

	// Status
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	statusText := m.status
	if m.paused {
		statusText = "⏸  PAUSED"
	}
	b.WriteString(statusStyle.Render(fmt.Sprintf("Status: %s", statusText)))
	b.WriteString("\n\n")

	// Progress bar
	progressBar := renderProgressBar(m.progress, 40)
	b.WriteString(progressBar)
	b.WriteString(fmt.Sprintf(" %.1f%%\n\n", m.progress*100))

	// Output section
	outputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(80)

	outputHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")).
		Render("Output:")

	b.WriteString(outputHeader)
	b.WriteString("\n")

	if len(m.output) == 0 {
		b.WriteString(outputStyle.Render("Waiting for output..."))
	} else {
		outputText := strings.Join(m.output, "\n")
		b.WriteString(outputStyle.Render(outputText))
	}

	b.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	if m.paused {
		b.WriteString(helpStyle.Render("P: Resume | S: Skip | Q: Quit"))
	} else {
		b.WriteString(helpStyle.Render("P: Pause | S: Skip | Q: Quit"))
	}

	return b.String()
}

// renderProgressBar renders a progress bar
func renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	return bar
}

// AddOutput adds a line to the output
func (m *ExecutionModel) AddOutput(line string) {
	m.output = append(m.output, line)
	if len(m.output) > m.maxOutputLines {
		m.output = m.output[len(m.output)-m.maxOutputLines:]
	}
}

// SetStatus sets the current status
func (m *ExecutionModel) SetStatus(status string) {
	m.status = status
}

// SetProgress sets the current progress
func (m *ExecutionModel) SetProgress(progress float64) {
	m.progress = progress
}

// SetCurrentPhase sets the current phase
func (m *ExecutionModel) SetCurrentPhase(phase string) {
	m.currentPhase = phase
}

// SetCurrentTask sets the current task
func (m *ExecutionModel) SetCurrentTask(task string) {
	m.currentTask = task
}
