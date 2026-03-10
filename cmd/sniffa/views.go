package main

import (
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	filewatcher "github.com/jwc20/sniffa/internal/filewatcher"
)

func (m model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	sidebarWidth := 30
	mainWidth := m.width - sidebarWidth - 3

	// Sidebar
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorTitle).
		BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).
		BorderForeground(colorBorder).Width(sidebarWidth-2).Padding(0, 1)

	sidebarInner := titleStyle.Render("sniffa") + "\n"
	sidebarInner += buildFileList(m.dirs, m.activeFile, sidebarWidth-2)
	sidebar := panelStyle.Width(sidebarWidth).Height(m.height - 2).Render(sidebarInner)

	// Main Content
	mainTitleStyle := titleStyle.Width(mainWidth - 2)
	mainInner := mainTitleStyle.Render("Test Output") + "\n" + buildOutput(m.output, mainWidth, m.height-4)
	main := panelStyle.Width(mainWidth).Height(m.height - 2).Render(mainInner)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)
	help := lipgloss.NewStyle().Foreground(colorMuted).Render("  q / ctrl+c  quit")

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, layout, help))
	v.AltScreen = true
	return v
}

func buildFileList(dirs []string, activeFile string, width int) string {
	itemStyle := lipgloss.NewStyle().Foreground(colorHighlight)
	activeStyle := lipgloss.NewStyle().Foreground(colorPass).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(colorMuted)

	// Use the utility to get the scoped directories
	toWatch := filewatcher.FindAllDirs(dirs, maxDepth)

	// pwd, _ := os.Getwd()
	var testFiles []string

	// Collect all _test.go files from the watched directories
	for _, dir := range toWatch {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
				// Construct the full relative path
				fullPath := filepath.Join(dir, e.Name())
				testFiles = append(testFiles, fullPath)
				// fmt.Println(fullPath)
				// relPath, err := filepath.Rel(pwd, fullPath)
				// fmt.Println(relPath)
				// if err == nil {
				// 	testFiles = append(testFiles, relPath)
				// }
			}
		}
	}

	if len(testFiles) == 0 {
		return mutedStyle.Render(" No tests found within depth...")
	}

	var sb strings.Builder
	for _, path := range testFiles {
		style := itemStyle

		// If the active path (from testResultMsg) matches or is contained in this path
		if activeFile != "" && strings.Contains(path, activeFile) {
			style = activeStyle
		}

		sb.WriteString(style.Render(path) + "\n")
	}

	return lipgloss.NewStyle().
		MaxWidth(width).
		Render(sb.String())
}
func buildOutput(lines []string, width, height int) string {
	passStyle := lipgloss.NewStyle().Foreground(colorPass).Bold(true)
	failStyle := lipgloss.NewStyle().Foreground(colorFail).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(colorMuted)

	if len(lines) == 0 {
		return mutedStyle.Render("\n  Watching for changes...")
	}

	var sb strings.Builder
	for _, block := range lines {
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "[PASS]"):
				sb.WriteString(passStyle.Render(line) + "\n")
			case strings.HasPrefix(line, "[FAIL]"):
				sb.WriteString(failStyle.Render(line) + "\n")
			case strings.HasPrefix(line, "--- FAIL"):
				sb.WriteString(failStyle.Render(line) + "\n")
			case strings.HasPrefix(line, "--- PASS"), strings.HasPrefix(line, "ok"):
				sb.WriteString(passStyle.Render(line) + "\n")
			default:
				sb.WriteString(line + "\n")
			}
		}
	}

	rendered := sb.String()
	allLines := strings.Split(rendered, "\n")
	if len(allLines) > height {
		allLines = allLines[len(allLines)-height:]
	}

	return lipgloss.NewStyle().
		MaxWidth(width).
		Render(strings.Join(allLines, "\n"))
}
