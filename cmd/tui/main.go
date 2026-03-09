package main

import (
	"log"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	altscreen bool
	width     int
	height    int
}

// ***************************************************************************

func main() {
	m := model{
		altscreen: true,
	}

	logfilePath := os.Getenv("BUBBLETEA_LOG")
	if logfilePath != "" {
		if _, err := tea.LogToFile(logfilePath, "simple"); err != nil {
			log.Fatal(err)
		}
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// ***************************************************************************

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+z":
			return m, tea.Suspend
		}
	}
	return m, nil
}

func (m model) center(content string) string {
	if m.width == 0 || m.height == 0 {
		return content
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	maxWidth := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}
	leftPad := (m.width - maxWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	prefix := strings.Repeat(" ", leftPad)
	topPad := (m.height - len(lines)) / 2
	if topPad < 0 {
		topPad = 0
	}
	var sb strings.Builder
	sb.WriteString(strings.Repeat("\n", topPad))
	for _, line := range lines {
		sb.WriteString(prefix + line + "\n")
	}
	return sb.String()
}

// ***************************************************************************

func (m model) View() tea.View {
	content := "Hello word"
	if m.altscreen {
		content = m.center(content)
	}

	v := tea.NewView(content)
	v.AltScreen = m.altscreen
	return v
}
