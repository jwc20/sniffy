package main

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	filewatcher "github.com/jwc20/sniffa/internal/filewatcher"
	scent "github.com/jwc20/sniffa/internal/scent"
)

type fileChangedMsg struct {
	pkg      string
	filename string
}

type testResultMsg struct {
	path   string
	output string
	passed bool
}

type Test struct {
	path    string
	pkg     string
	names   []string
	runner  *scent.RunnerConfig
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
	changes chan fileChangedMsg
	scent   *scent.Scent
}

var testFuncRe = regexp.MustCompile(`^func (Test\w+)\(`)

func parseTestNames(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := testFuncRe.FindStringSubmatch(scanner.Text()); m != nil {
			names = append(names, m[1])
		}
	}
	return names
}

func scentExtensions(s *scent.Scent) []string {
	if s == nil {
		return []string{".go"}
	}
	seen := map[string]bool{}
	var exts []string
	for _, v := range s.Validators {
		for _, e := range v.Extensions {
			ext := strings.ToLower(e)
			if !seen[ext] {
				seen[ext] = true
				exts = append(exts, ext)
			}
		}
	}
	if len(exts) == 0 {
		return []string{".go"}
	}
	return exts
}

func initTests(dirs []string, s *scent.Scent) []Test {
	exts := scentExtensions(s)
	toWatch := filewatcher.FindAllDirs(dirs, maxDepth, exts)
	var tests []Test
	for _, dir := range toWatch {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			for _, ext := range exts {
				if !strings.HasSuffix(name, ext) {
					continue
				}
				fullPath := filepath.Join(dir, name)
				if ext == ".go" {
					if !strings.HasSuffix(name, "_test.go") {
						continue
					}
				} else if s != nil {
					if !s.IsTestFile(fullPath) {
						continue
					}
				} else {
					continue
				}
				t := Test{
					path:    fullPath,
					pkg:     filepath.Clean(dir),
					enabled: true,
				}
				if ext == ".go" {
					t.names = parseTestNames(fullPath)
				} else if s != nil {
					t.runner = s.RunnerForFile(fullPath)
				}
				tests = append(tests, t)
				break
			}
		}
	}
	return tests
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		startWatcher(m.dirs, m.changes),
		listenForChange(m.changes),
		listenForResult(m.results),
		runAllTestsCmd(m.tests, m.results),
	)
}

func startWatcher(dirs []string, ch chan fileChangedMsg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			ctx := context.Background()
			filewatcher.Watch(ctx, dirs, func(event filewatcher.Event) error {
				ch <- fileChangedMsg{pkg: filepath.Clean(event.PkgPath), filename: event.Filename}
				return nil
			})
		}()
		return nil
	}
}

func listenForChange(ch chan fileChangedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func listenForResult(ch chan testResultMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func runTestCmd(t Test, results chan testResultMsg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			var out string
			var passed bool
			if t.runner != nil {
				out, passed = t.runner.Execute()
			} else {
				out, passed = runGoTests(t.pkg, t.names)
			}
			results <- testResultMsg{path: t.path, output: out, passed: passed}
		}()
		return nil
	}
}

func runAllTestsCmd(tests []Test, results chan testResultMsg) tea.Cmd {
	var cmds []tea.Cmd
	for _, t := range tests {
		t := t
		if t.enabled {
			cmds = append(cmds, runTestCmd(t, results))
		}
	}
	return tea.Batch(cmds...)
}

func runGoTests(pkgPath string, names []string) (string, bool) {
	args := []string{"test", "-v"}
	if len(names) > 0 {
		args = append(args, "-run", "^("+strings.Join(names, "|")+")$")
	}
	args = append(args, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = pkgPath
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

func (m model) testsForPkg(pkg string) []Test {
	var out []Test
	for _, t := range m.tests {
		if t.pkg == pkg {
			out = append(out, t)
		}
	}
	return out
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case fileChangedMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, listenForChange(m.changes))
		for _, t := range m.testsForPkg(msg.pkg) {
			if t.enabled {
				cmds = append(cmds, runTestCmd(t, m.results))
			}
		}
		return m, tea.Batch(cmds...)

	case testResultMsg:
		for i, t := range m.tests {
			if t.path == msg.path {
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
		case "space":
			if len(m.tests) > 0 {
				t := &m.tests[m.cursor]
				t.enabled = !t.enabled
				if t.enabled {
					return m, runTestCmd(*t, m.results)
				}
			}
		}
	}
	return m, nil
}
