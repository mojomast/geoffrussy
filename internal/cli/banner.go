package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// Banner returns the Geoffrussy ASCII art banner with a horizontal color gradient
func Banner() string {
	bannerText := `
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
`

	// Define gradient colors: cyan (#3CADFF) to purple (#BA3CFF)
	// Using hardcoded valid hex values, so errors can be safely ignored
	startColor, _ := colorful.Hex("#3CADFF")
	endColor, _ := colorful.Hex("#BA3CFF")

	// Split banner into lines
	lines := strings.Split(bannerText, "\n")

	// Find the maximum line width for gradient calculation
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Build the gradient banner
	var result strings.Builder

	for lineIdx, line := range lines {
		for i, char := range line {
			if char == ' ' {
				// Don't color spaces, just add them
				result.WriteRune(char)
			} else {
				// Calculate gradient position based on character position relative to max width
				t := float64(i) / float64(maxWidth-1)
				if maxWidth <= 1 {
					t = 0
				}

				// Blend colors
				gradientColor := startColor.BlendLuv(endColor, t)

				// Apply color to character and reset immediately after
				coloredChar := lipgloss.NewStyle().
					Foreground(lipgloss.Color(gradientColor.Hex())).
					Render(string(char))

				result.WriteString(coloredChar)
			}
		}
		// Add newline after each line except the last to avoid extra trailing newlines
		if lineIdx < len(lines)-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}
