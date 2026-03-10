package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	filewatcher "github.com/jwc20/sniffa/internal/filewatcher"
)

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
	return watchCmd(m.dirs)
}

func watchCmd(dirs []string) tea.Cmd {
	ch := make(chan testResultMsg)
	return tea.Batch(func() tea.Msg {
		go func() {
			ctx := context.Background()
			filewatcher.Watch(ctx, dirs, false, func(event filewatcher.Event) error {
				out, passed := runTests(event.PkgPath)
				ch <- testResultMsg{pkg: event.PkgPath, output: out, passed: passed}
				return nil
			})
		}()
		return nil
	}, func() tea.Msg {
		return <-ch
	})
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
		pwd, _ := os.Getwd()
		relPkg, _ := filepath.Rel(pwd, msg.pkg)
		m.activeFile = relPkg
		return m, watchCmd(m.dirs)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}
