package main

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	sidebarWidth := 30
	mainWidth := m.width - sidebarWidth - 3

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.styles.Title).
		BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).
		BorderForeground(m.styles.Border).Width(sidebarWidth-2).Padding(0, 1)

	sidebarInner := titleStyle.Render("sniffy") + "\n"
	sidebarInner += buildFileList(m.tests, m.cursor, sidebarWidth-2, m.styles)
	sidebar := m.styles.Panel.Width(sidebarWidth).Height(m.height - 2).Render(sidebarInner)

	mainTitleStyle := titleStyle.Width(mainWidth - 2)

	var hoveredResult *testResultMsg
	if len(m.tests) > 0 {
		hoveredResult = m.tests[m.cursor].result
	}

	mainInner := mainTitleStyle.Render("Test Output") + "\n" + buildOutput(hoveredResult, mainWidth, m.height-4, m.styles)
	main := m.styles.Panel.Width(mainWidth).Height(m.height - 2).Render(mainInner)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)
	help := lipgloss.NewStyle().Foreground(m.styles.Muted).Render("  ↑/↓  navigate    space  toggle    q  quit")

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, layout, help))
	v.AltScreen = true
	return v
}

func testColor(t Test, styles Styles) color.Color {
	if !t.enabled {
		return styles.Muted
	}
	if t.result == nil {
		return styles.Highlight
	}
	if t.result.passed {
		return styles.Pass
	}
	return styles.Fail
}

func lighten(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	blend := func(ch uint32) uint8 {
		v := float64(ch>>8) / 255.0
		v = v + (1.0-v)*0.35
		if v > 1.0 {
			v = 1.0
		}
		return uint8(v * 255)
	}
	return color.RGBA{R: blend(r), G: blend(g), B: blend(b), A: uint8(a >> 8)}
}

func buildFileList(tests []Test, cursor int, width int, styles Styles) string {
	mutedStyle := lipgloss.NewStyle().Foreground(styles.Muted)

	if len(tests) == 0 {
		return mutedStyle.Render(" No tests found within depth...")
	}

	var sb strings.Builder
	for i, t := range tests {
		isCursor := i == cursor

		base := testColor(t, styles)
		c := base
		if isCursor {
			c = lighten(base)
		}

		style := lipgloss.NewStyle().Foreground(c)
		if isCursor {
			style = style.Bold(true)
		}

		line := style.Render(t.path)
		sb.WriteString(line + "\n")
	}

	return lipgloss.NewStyle().MaxWidth(width).Render(sb.String())
}

func buildOutput(result *testResultMsg, width, height int, styles Styles) string {
	passStyle := lipgloss.NewStyle().Foreground(styles.Pass).Bold(true)
	failStyle := lipgloss.NewStyle().Foreground(styles.Fail).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.Muted)

	if result == nil {
		return mutedStyle.Render("\n  Watching for changes...")
	}

	var sb strings.Builder
	for _, line := range strings.Split(result.output, "\n") {
		switch {
		case strings.HasPrefix(line, "=== RUN"):
			continue
		case strings.Contains(line, "FAIL"):
			sb.WriteString(failStyle.Render(line) + "\n")
		case strings.Contains(line, "PASS") || strings.Contains(line, "ok") || strings.Contains(line, "passed"):
			sb.WriteString(passStyle.Render(line) + "\n")
		default:
			sb.WriteString(line + "\n")
		}
	}

	allLines := strings.Split(sb.String(), "\n")
	if len(allLines) > height {
		allLines = allLines[len(allLines)-height:]
	}

	return lipgloss.NewStyle().MaxWidth(width).Render(strings.Join(allLines, "\n"))
}
