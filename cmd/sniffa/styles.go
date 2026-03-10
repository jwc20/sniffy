package main

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

var (
	hasDarkBG bool
	lightDark lipgloss.LightDarkFunc

	colorBorder    color.Color
	colorTitle     color.Color
	colorPass      color.Color
	colorFail      color.Color
	colorMuted     color.Color
	colorHighlight color.Color

	panelStyle     lipgloss.Style
	titleStyle     lipgloss.Style
	mainTitleStyle lipgloss.Style
)

func initColors() {
	hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDarkBG)

	colorBorder = lightDark(lipgloss.Color("#C8C8C8"), lipgloss.Color("#444444"))
	colorTitle = lightDark(lipgloss.Color("#1A1A1A"), lipgloss.Color("#E8E8E8"))
	colorPass = lightDark(lipgloss.Color("#2E7D32"), lipgloss.Color("#66BB6A"))
	colorFail = lightDark(lipgloss.Color("#C62828"), lipgloss.Color("#EF5350"))
	colorMuted = lightDark(lipgloss.Color("#757575"), lipgloss.Color("#757575"))
	colorHighlight = lightDark(lipgloss.Color("#1565C0"), lipgloss.Color("#42A5F5"))

	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder)
}
