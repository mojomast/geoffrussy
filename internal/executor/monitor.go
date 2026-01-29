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
	executor     *Executor
	projectID    string
	width        int
	height       int
	viewport     viewport.Model
	progress     progress.Model
	updates      []TaskUpdate
	currentTask  string
	currentPhase string
	startTime    time.Time
	completion   float64
	phasesDone   int
	totalPhases  int
	tasksDone    int
	totalTasks   int
	tokensIn     int
	tokensOut    int
	requests     int
	err          error
}

// NewMonitor creates a new live monitor
func NewMonitor(executor *Executor, projectID string) *Monitor {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 0, 1, 1)

	prog := progress.New(progress.WithDefaultGradient())

	return &Monitor{
		executor:  executor,
		projectID: projectID,
		viewport:  vp,
		progress:  prog,
		updates:   []TaskUpdate{},
		startTime: time.Now(),
	}
}

// monitorMsg is a message containing a task update
type monitorMsg TaskUpdate

// tickMsg is sent every second to update the timer
type tickMsg time.Time

// Init initializes the monitor
func (m *Monitor) Init() tea.Cmd {
	return tea.Batch(
		m.waitForUpdate(),
		m.progress.Init(),
		tickCmd(),
	)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
		m.viewport.Height = msg.Height - 25

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

	case tickMsg:
		m.refreshStats()
		cmds = append(cmds, tickCmd())
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

	// Header - Banner
	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(bannerStyle.Render(`
 /$$$$$$                       /$$$$$$   /$$$$$$
 /$$__  $$                     /$$__  $$ /$$__  $$
| $$  \__/  /$$$$$$   /$$$$$$ | $$  \__/| $$  \__//$$$$$$  /$$   /$$  /$$$$$$$ /$$$$$$$ /$$   /$$
| $$ /$$$$ /$$__  $$ /$$__  $$| $$$$    | $$$$   /$$__  $$| $$  | $$ /$$_____//$$_____/| $$  | $$
| $$|_  $$| $$$$$$$$| $$  \ $$| $$_/    | $$_/  | $$  \__/| $$  | $$|  $$$$$$|  $$$$$$ | $$  | $$
| $$  \ $$| $$_____/| $$  | $$| $$      | $$    | $$      | $$  | $$ \____  $$\____  $$| $$  | $$
|  $$$$$$/|  $$$$$$$|  $$$$$$/| $$      | $$    | $$      |  $$$$$$/ /$$$$$$$//$$$$$$$/|  $$$$$$$
 \______/  \_______/ \______/ |__/      |__/    |__/       \______/ |_______/|_______/  \____  $$
                                                                                        /$$  | $$
                                                                                       |  $$$$$$/
                                                                                        \______/
`))

	b.WriteString("\n")

	// Stats row
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.totalTasks > 0 {
		completionStr := fmt.Sprintf("%.0f%%", m.completion)
		phaseStr := fmt.Sprintf("%d/%d", m.phasesDone, m.totalPhases)
		taskStr := fmt.Sprintf("%d/%d", m.tasksDone, m.totalTasks)
		stats := fmt.Sprintf("%s tasks • %s phases • %s done", taskStr, phaseStr, completionStr)
		b.WriteString(statsStyle.Render(stats))
	}

	// Token stats
	if m.tokensIn > 0 || m.tokensOut > 0 {
		tokenStr := fmt.Sprintf("  |  🔤 In: %d  Out: %d", m.tokensIn, m.tokensOut)
		b.WriteString(statsStyle.Render(tokenStr))
	}
	b.WriteString("\n")

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

	// Progress bar
	if m.totalTasks > 0 {
		m.progress.SetPercent(m.completion / 100)
		b.WriteString(m.progress.View())
		b.WriteString("\n\n")
	}

	// Output viewport - this is where task updates appear
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

	// Show last 30 updates (fewer to keep UI cleaner)
	start := 0
	if len(m.updates) > 30 {
		start = len(m.updates) - 30
	}

	for _, update := range m.updates[start:] {
		line := m.formatUpdate(update)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
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

// refreshStats refreshes statistics from the store
func (m *Monitor) refreshStats() {
	stats, err := m.executor.store.CalculateProgress(m.projectID)
	if err != nil {
		return
	}

	m.completion = stats.CompletionPercentage
	m.phasesDone = stats.CompletedPhases
	m.totalPhases = stats.TotalPhases
	m.tasksDone = stats.CompletedTasks
	m.totalTasks = stats.TotalTasks

	tokenStats, err := m.executor.store.GetTokenStats(m.projectID)
	if err != nil {
		return
	}

	m.tokensIn = tokenStats.TotalInput
	m.tokensOut = tokenStats.TotalOutput
}

// RunWithOutput runs the monitor and writes output to the given writer
func (m *Monitor) RunWithOutput(w io.Writer) error {
	p := tea.NewProgram(m, tea.WithOutput(w), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running monitor: %w", err)
	}
	return nil
}
