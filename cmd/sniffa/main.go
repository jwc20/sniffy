package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
	filewatcher "github.com/jwc20/sniffa/internal/filewatcher"
)

const maxDepth = 7

var (
	hasDarkBG bool
	lightDark lipgloss.LightDarkFunc

	colorBorder    color.Color
	colorTitle     color.Color
	colorPass      color.Color
	colorFail      color.Color
	colorMuted     color.Color
	colorHighlight color.Color
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
}

type testResultMsg struct {
	pkg    string
	output string
	passed bool
}

type model struct {
	width      int
	height     int
	dirs       []string
	output     []string
	activeFile string
	cancel     context.CancelFunc
}

func (m model) Init() tea.Cmd {
	return startWatcher(m.dirs)
}

func startWatcher(dirs []string) tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func watchCmd(dirs []string) tea.Cmd {
	ch := make(chan testResultMsg)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		_ = cancel
		filewatcher.Watch(ctx, dirs, false, func(event filewatcher.Event) error {
			out, passed := runTests(event.PkgPath)
			ch <- testResultMsg{pkg: event.PkgPath, output: out, passed: passed}
			return nil
		})
	}()
	return func() tea.Msg {
		return <-ch
	}
}

func runTests(pkgPath string) (string, bool) {
	cmd := exec.Command("go", "test", "-v", pkgPath)
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if len(m.output) == 0 {
			return m, watchCmd(m.dirs)
		}
		return m, nil

	case testResultMsg:
		status := "PASS"
		if !msg.passed {
			status = "FAIL"
		}
		header := fmt.Sprintf("[%s] %s\n%s\n", status, msg.pkg, strings.Repeat("-", 40))
		m.output = append(m.output, header+msg.output)
		if len(m.output) > 100 {
			m.output = m.output[len(m.output)-100:]
		}
		m.activeFile = msg.pkg
		return m, watchCmd(m.dirs)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	sidebarWidth := 30
	mainWidth := m.width - sidebarWidth - 3

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorTitle).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorBorder).
		Width(sidebarWidth-2).
		Padding(0, 1)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder)

	sidebarInner := titleStyle.Render("sniffa") + "\n"
	sidebarInner += buildTree(m.dirs, m.activeFile, sidebarWidth-2)

	sidebar := panelStyle.
		Width(sidebarWidth).
		Height(m.height - 2).
		Render(sidebarInner)

	outputContent := buildOutput(m.output, mainWidth, m.height-4)

	mainTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorTitle).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorBorder).
		Width(mainWidth-2).
		Padding(0, 1)

	mainInner := mainTitleStyle.Render("Test Output") + "\n" + outputContent

	main := panelStyle.
		Width(mainWidth).
		Height(m.height - 2).
		Render(mainInner)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)

	helpStyle := lipgloss.NewStyle().Foreground(colorMuted)
	help := helpStyle.Render("  q / ctrl+c  quit")

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, layout, help))
	v.AltScreen = true
	return v
}

func buildTree(dirs []string, activeFile string, width int) string {
	enumeratorStyle := lipgloss.NewStyle().Foreground(colorMuted).PaddingRight(1)
	itemStyle := lipgloss.NewStyle().Foreground(colorHighlight)
	activeStyle := lipgloss.NewStyle().Foreground(colorPass).Bold(true)

	toWatch := filewatcher.FindAllDirs(dirs, maxDepth)

	pwd, _ := os.Getwd()
	base := filepath.Base(pwd)

	root := tree.Root(base).
		IndenterStyle(enumeratorStyle).
		EnumeratorStyle(enumeratorStyle).
		RootStyle(lipgloss.NewStyle().Foreground(colorTitle).Bold(true)).
		ItemStyle(itemStyle)

	for _, dir := range toWatch {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		relDir, _ := filepath.Rel(pwd, dir)
		if relDir == "." {
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), "_test.go") {
					name := e.Name()
					style := itemStyle
					if strings.Contains(activeFile, strings.TrimSuffix(name, "_test.go")) {
						style = activeStyle
					}
					root.Child(style.Render(name))
				}
			}
		} else {
			branch := tree.Root(relDir)
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), "_test.go") {
					name := e.Name()
					style := itemStyle
					if strings.Contains(activeFile, strings.TrimSuffix(name, "_test.go")) {
						style = activeStyle
					}
					branch.Child(style.Render(name))
				}
			}
			root.Child(branch)
		}
	}

	return lipgloss.NewStyle().
		MaxWidth(width).
		Render(root.String())
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

func main() {
	initColors()

	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}

	m := model{dirs: dirs}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
