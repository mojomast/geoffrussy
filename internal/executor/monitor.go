package executor

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Monitor provides live monitoring of task execution
type Monitor struct {
	executor   *Executor
	width      int
	height     int
	viewport   viewport.Model
	progress   progress.Model
	updates    []TaskUpdate
	currentTask string
	currentPhase string
	startTime   time.Time
	err         error
}

// NewMonitor creates a new live monitor
func NewMonitor(executor *Executor) *Monitor {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1)

	prog := progress.New(progress.WithDefaultGradient())

	return &Monitor{
		executor:  executor,
		viewport:  vp,
		progress:  prog,
		updates:   []TaskUpdate{},
		startTime: time.Now(),
	}
}

// monitorMsg is a message containing a task update
type monitorMsg TaskUpdate

// Init initializes the monitor
func (m *Monitor) Init() tea.Cmd {
	return tea.Batch(
		m.waitForUpdate(),
		m.progress.Init(),
	)
}

// Update handles messages and updates the monitor
func (m *Monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.executor.Close()
			return m, tea.Quit

		case "p":
			if err := m.executor.PauseExecution(); err != nil {
				m.err = err
			}
			return m, nil

		case "r":
			if err := m.executor.ResumeExecution(); err != nil {
				m.err = err
			}
			return m, nil

		case "s":
			if m.currentTask != "" {
				if err := m.executor.SkipTask(m.currentTask); err != nil {
					m.err = err
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 15

	case monitorMsg:
		update := TaskUpdate(msg)
		m.updates = append(m.updates, update)

		// Update current task and phase
		if update.TaskID != "" {
			m.currentTask = update.TaskID
		}
		if update.PhaseID != "" {
			m.currentPhase = update.PhaseID
		}

		// Update viewport content
		m.updateViewport()

		// Wait for next update
		return m, m.waitForUpdate()

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the monitor UI
func (m *Monitor) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("🚀 Geoffrey Execution Monitor"))
	b.WriteString("\n\n")

	// Current phase and task
	if m.currentPhase != "" {
		phaseStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))
		b.WriteString(phaseStyle.Render(fmt.Sprintf("Phase: %s", m.currentPhase)))
		b.WriteString("\n")
	}

	if m.currentTask != "" {
		taskStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))
		b.WriteString(taskStyle.Render(fmt.Sprintf("Task: %s", m.currentTask)))
		b.WriteString("\n")
	}

	// Elapsed time
	elapsed := time.Since(m.startTime)
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))
	b.WriteString(timeStyle.Render(fmt.Sprintf("Elapsed: %s", formatDuration(elapsed))))
	b.WriteString("\n\n")

	// Progress bar (simplified - would need actual progress calculation)
	b.WriteString(m.progress.View())
	b.WriteString("\n\n")

	// Output viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render("P: Pause | R: Resume | S: Skip | Q: Quit"))

	return b.String()
}

// waitForUpdate waits for the next task update
func (m *Monitor) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		select {
		case update := <-m.executor.StreamOutput():
			return monitorMsg(update)
		}
	}
}

// updateViewport updates the viewport content with recent updates
func (m *Monitor) updateViewport() {
	var lines []string

	// Show last 50 updates
	start := 0
	if len(m.updates) > 50 {
		start = len(m.updates) - 50
	}

	for _, update := range m.updates[start:] {
		line := m.formatUpdate(update)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)

	// Scroll to bottom
	m.viewport.GotoBottom()
}

// formatUpdate formats a task update for display
func (m *Monitor) formatUpdate(update TaskUpdate) string {
	timestamp := update.Timestamp.Format("15:04:05")

	var icon string
	var color lipgloss.Color

	switch update.Type {
	case TaskStarted:
		icon = "▶"
		color = lipgloss.Color("86")
	case TaskProgress:
		icon = "⋯"
		color = lipgloss.Color("141")
	case TaskCompleted:
		icon = "✓"
		color = lipgloss.Color("82")
	case TaskError:
		icon = "✗"
		color = lipgloss.Color("196")
	case TaskBlocked:
		icon = "⚠"
		color = lipgloss.Color("226")
	case TaskPaused:
		icon = "⏸"
		color = lipgloss.Color("226")
	case TaskResumed:
		icon = "▶"
		color = lipgloss.Color("86")
	case TaskSkipped:
		icon = "⏭"
		color = lipgloss.Color("241")
	default:
		icon = "•"
		color = lipgloss.Color("241")
	}

	style := lipgloss.NewStyle().Foreground(color)
	timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	return fmt.Sprintf("%s %s %s",
		timestampStyle.Render(timestamp),
		style.Render(icon),
		update.Content)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// Run runs the monitor as a Bubbletea program
func (m *Monitor) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running monitor: %w", err)
	}
	return nil
}

// RunWithOutput runs the monitor and writes output to the given writer
func (m *Monitor) RunWithOutput(w io.Writer) error {
	p := tea.NewProgram(m, tea.WithOutput(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running monitor: %w", err)
	}
	return nil
}
