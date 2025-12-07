package comments

import (
	"tuistr/components/colors"

	"github.com/charmbracelet/lipgloss"
)

var viewportStyle = lipgloss.NewStyle().Margin(0, 2, 1, 2)

var (
	commentAuthorStyle = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Blue)).Bold(true)
	commentDateStyle   = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Lavender)).Italic(true)
	commentTextStyle   = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Text))
	collapsedStyle     = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Yellow))
)

var (
	postAuthorStyle    = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Blue))
	postTextStyle      = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Sand))
	postTimestampStyle = lipgloss.NewStyle().Foreground(colors.AdaptiveColor(colors.Text)).Faint(true)
)
