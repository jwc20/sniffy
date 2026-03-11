package main

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

type Styles struct {
	Border    color.Color
	Title     color.Color
	Pass      color.Color
	Fail      color.Color
	Muted     color.Color
	Highlight color.Color
	Cursor    color.Color
	Panel     lipgloss.Style
}

func DefaultStyles() Styles {
	hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(hasDarkBG)

	s := Styles{
		Border:    lightDark(lipgloss.Color("#C8C8C8"), lipgloss.Color("#444444")),
		Title:     lightDark(lipgloss.Color("#1A1A1A"), lipgloss.Color("#E8E8E8")),
		Pass:      lightDark(lipgloss.Color("#2E7D32"), lipgloss.Color("#66BB6A")),
		Fail:      lightDark(lipgloss.Color("#C62828"), lipgloss.Color("#EF5350")),
		Muted:     lightDark(lipgloss.Color("#757575"), lipgloss.Color("#757575")),
		Highlight: lightDark(lipgloss.Color("#1565C0"), lipgloss.Color("#42A5F5")),
		Cursor:    lightDark(lipgloss.Color("#F57F17"), lipgloss.Color("#FFD54F")),
	}

	s.Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border)

	return s
}
