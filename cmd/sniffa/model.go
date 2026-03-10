package main

import (
	"context"
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

type Test struct {
	path    string
	pkg     string
	enabled bool
	result  *testResultMsg
}

type model struct {
	width   int
	height  int
	dirs    []string
	tests   []Test
	cursor  int
	results chan testResultMsg
}

func initTests(dirs []string) []Test {
	toWatch := filewatcher.FindAllDirs(dirs, maxDepth)
	var tests []Test
	for _, dir := range toWatch {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
				tests = append(tests, Test{
					path:    filepath.Join(dir, e.Name()),
					pkg:     filepath.Clean(dir),
					enabled: true,
				})
			}
		}
	}
	return tests
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		startWatcher(m.dirs, m.results),
		listenForResult(m.results),
		runAllTestsCmd(m.tests),
	)
}

func startWatcher(dirs []string, ch chan testResultMsg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			ctx := context.Background()
			filewatcher.Watch(ctx, dirs, false, func(event filewatcher.Event) error {
				out, passed := runTests(event.PkgPath)
				ch <- testResultMsg{pkg: filepath.Clean(event.PkgPath), output: out, passed: passed}
				return nil
			})
		}()
		return nil
	}
}

func listenForResult(ch chan testResultMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func runAllTestsCmd(tests []Test) tea.Cmd {
	var cmds []tea.Cmd
	for _, t := range tests {
		t := t
		if t.enabled {
			cmds = append(cmds, func() tea.Msg {
				out, passed := runTests(t.pkg)
				return testResultMsg{pkg: t.pkg, output: out, passed: passed}
			})
		}
	}
	return tea.Batch(cmds...)
}

func runTests(pkgPath string) (string, bool) {
	cmd := exec.Command("go", "test", "-v", ".")
	cmd.Dir = pkgPath
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case testResultMsg:
		for i, t := range m.tests {
			if filepath.Clean(t.pkg) == filepath.Clean(msg.pkg) {
				result := msg
				m.tests[i].result = &result
			}
		}
		return m, listenForResult(m.results)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tests)-1 {
				m.cursor++
			}
		case " ", "enter":
			if len(m.tests) > 0 {
				m.tests[m.cursor].enabled = !m.tests[m.cursor].enabled
			}
		}
	}
	return m, nil
}
