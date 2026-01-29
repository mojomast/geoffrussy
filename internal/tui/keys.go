package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap defines the keyboard shortcuts for the TUI
type KeyMap struct {
	Up          []string
	Down        []string
	Left        []string
	Right       []string
	Enter       []string
	Space       []string
	Escape      []string
	Quit        []string
	Pause       []string
	Skip        []string
	SelectAll   []string
	Help        []string
	PageUp      []string
	PageDown    []string
	Home        []string
	End         []string
}

// DefaultKeyMap returns the default key mappings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        []string{"up", "k"},
		Down:      []string{"down", "j"},
		Left:      []string{"left", "h"},
		Right:     []string{"right", "l"},
		Enter:     []string{"enter"},
		Space:     []string{" ", "space"},
		Escape:    []string{"esc"},
		Quit:      []string{"q", "ctrl+c"},
		Pause:     []string{"p", "ctrl+p"},
		Skip:      []string{"s"},
		SelectAll: []string{"a"},
		Help:      []string{"?"},
		PageUp:    []string{"pgup"},
		PageDown:  []string{"pgdown"},
		Home:      []string{"home"},
		End:       []string{"end"},
	}
}

// Matches checks if a key message matches any of the key bindings
func (k KeyMap) Matches(msg tea.KeyMsg, keys []string) bool {
	for _, key := range keys {
		if msg.String() == key {
			return true
		}
	}
	return false
}

// IsUp checks if the key is an up navigation key
func (k KeyMap) IsUp(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Up)
}

// IsDown checks if the key is a down navigation key
func (k KeyMap) IsDown(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Down)
}

// IsLeft checks if the key is a left navigation key
func (k KeyMap) IsLeft(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Left)
}

// IsRight checks if the key is a right navigation key
func (k KeyMap) IsRight(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Right)
}

// IsEnter checks if the key is an enter key
func (k KeyMap) IsEnter(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Enter)
}

// IsSpace checks if the key is a space key
func (k KeyMap) IsSpace(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Space)
}

// IsEscape checks if the key is an escape key
func (k KeyMap) IsEscape(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Escape)
}

// IsQuit checks if the key is a quit key
func (k KeyMap) IsQuit(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Quit)
}

// IsPause checks if the key is a pause key
func (k KeyMap) IsPause(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Pause)
}

// IsSkip checks if the key is a skip key
func (k KeyMap) IsSkip(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Skip)
}

// IsSelectAll checks if the key is a select all key
func (k KeyMap) IsSelectAll(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.SelectAll)
}

// IsHelp checks if the key is a help key
func (k KeyMap) IsHelp(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Help)
}

// IsPageUp checks if the key is a page up key
func (k KeyMap) IsPageUp(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.PageUp)
}

// IsPageDown checks if the key is a page down key
func (k KeyMap) IsPageDown(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.PageDown)
}

// IsHome checks if the key is a home key
func (k KeyMap) IsHome(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.Home)
}

// IsEnd checks if the key is an end key
func (k KeyMap) IsEnd(msg tea.KeyMsg) bool {
	return k.Matches(msg, k.End)
}

// Viewport represents a scrollable viewport
type Viewport struct {
	Width  int
	Height int
	YOffset int
	Content []string
}

// NewViewport creates a new viewport
func NewViewport(width, height int) Viewport {
	return Viewport{
		Width:   width,
		Height:  height,
		YOffset: 0,
		Content: []string{},
	}
}

// SetContent sets the viewport content
func (v *Viewport) SetContent(content []string) {
	v.Content = content
}

// ScrollUp scrolls the viewport up
func (v *Viewport) ScrollUp(lines int) {
	v.YOffset -= lines
	if v.YOffset < 0 {
		v.YOffset = 0
	}
}

// ScrollDown scrolls the viewport down
func (v *Viewport) ScrollDown(lines int) {
	maxOffset := len(v.Content) - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}

	v.YOffset += lines
	if v.YOffset > maxOffset {
		v.YOffset = maxOffset
	}
}

// ScrollToTop scrolls to the top
func (v *Viewport) ScrollToTop() {
	v.YOffset = 0
}

// ScrollToBottom scrolls to the bottom
func (v *Viewport) ScrollToBottom() {
	maxOffset := len(v.Content) - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	v.YOffset = maxOffset
}

// PageUp scrolls up one page
func (v *Viewport) PageUp() {
	v.ScrollUp(v.Height)
}

// PageDown scrolls down one page
func (v *Viewport) PageDown() {
	v.ScrollDown(v.Height)
}

// View returns the visible content
func (v *Viewport) View() string {
	if len(v.Content) == 0 {
		return ""
	}

	start := v.YOffset
	end := v.YOffset + v.Height

	if start >= len(v.Content) {
		start = len(v.Content) - 1
	}
	if end > len(v.Content) {
		end = len(v.Content)
	}

	visible := v.Content[start:end]
	return joinLines(visible)
}

// joinLines joins lines with newlines
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

// HandleResize handles terminal resize events
func HandleResize(width, height int) tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  width,
			Height: height,
		}
	}
}
