package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusModel represents the TUI model for project status dashboard
type StatusModel struct {
	projectName    string
	currentStage   string
	currentPhase   string
	phasesTotal    int
	phasesComplete int
	phasesProgress int
	phasesPending  int
	blockers       []string
	totalTokens    int
	totalCost      float64
	err            error
	quitting       bool
}

// NewStatusModel creates a new status TUI model
func NewStatusModel() StatusModel {
	return StatusModel{
		blockers: []string{},
	}
}

// Init initializes the status model
func (m StatusModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the status dashboard UI
func (m StatusModel) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("📊 Project Status Dashboard"))
	b.WriteString("\n\n")

	// Project info section
	projectStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(80)

	var projectInfo strings.Builder
	projectInfo.WriteString(fmt.Sprintf("📁 Project: %s\n", m.projectName))
	projectInfo.WriteString(fmt.Sprintf("🎯 Stage: %s\n", m.currentStage))
	if m.currentPhase != "" {
		projectInfo.WriteString(fmt.Sprintf("📋 Current Phase: %s\n", m.currentPhase))
	}

	b.WriteString(projectStyle.Render(projectInfo.String()))
	b.WriteString("\n\n")

	// Phase progress section
	progressStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	b.WriteString(progressStyle.Render("📈 Phase Progress"))
	b.WriteString("\n\n")

	// Progress stats
	statsStyle := lipgloss.NewStyle().
		Padding(0, 2)

	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("✅ Completed:    %d\n", m.phasesComplete))
	stats.WriteString(fmt.Sprintf("🔄 In Progress:  %d\n", m.phasesProgress))
	stats.WriteString(fmt.Sprintf("⏳ Not Started:  %d\n", m.phasesPending))
	stats.WriteString(fmt.Sprintf("📊 Total:        %d\n", m.phasesTotal))

	b.WriteString(statsStyle.Render(stats.String()))
	b.WriteString("\n")

	// Progress bar
	if m.phasesTotal > 0 {
		progress := float64(m.phasesComplete) / float64(m.phasesTotal)
		progressBar := renderProgressBar(progress, 60)
		b.WriteString(fmt.Sprintf("  %s %.1f%%\n", progressBar, progress*100))
	}

	b.WriteString("\n")

	// Blockers section
	if len(m.blockers) > 0 {
		blockerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

		b.WriteString(blockerStyle.Render("🚫 Active Blockers"))
		b.WriteString("\n\n")

		blockerListStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1).
			Width(80)

		var blockerList strings.Builder
		for i, blocker := range m.blockers {
			blockerList.WriteString(fmt.Sprintf("%d. %s\n", i+1, blocker))
		}

		b.WriteString(blockerListStyle.Render(blockerList.String()))
		b.WriteString("\n\n")
	}

	// Token usage section
	tokenStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141"))

	b.WriteString(tokenStyle.Render("💰 Token Usage"))
	b.WriteString("\n\n")

	tokenStatsStyle := lipgloss.NewStyle().
		Padding(0, 2)

	var tokenStats strings.Builder
	tokenStats.WriteString(fmt.Sprintf("Total Tokens: %d\n", m.totalTokens))
	tokenStats.WriteString(fmt.Sprintf("Total Cost: $%.4f\n", m.totalCost))

	b.WriteString(tokenStatsStyle.Render(tokenStats.String()))
	b.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	b.WriteString(helpStyle.Render("Q: Quit"))

	return b.String()
}

// SetProjectInfo sets the project information
func (m *StatusModel) SetProjectInfo(name, stage, phase string) {
	m.projectName = name
	m.currentStage = stage
	m.currentPhase = phase
}

// SetPhaseProgress sets the phase progress information
func (m *StatusModel) SetPhaseProgress(total, complete, progress, pending int) {
	m.phasesTotal = total
	m.phasesComplete = complete
	m.phasesProgress = progress
	m.phasesPending = pending
}

// SetBlockers sets the list of active blockers
func (m *StatusModel) SetBlockers(blockers []string) {
	m.blockers = blockers
}

// SetTokenUsage sets the token usage information
func (m *StatusModel) SetTokenUsage(tokens int, cost float64) {
	m.totalTokens = tokens
	m.totalCost = cost
}
