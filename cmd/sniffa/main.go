package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	scent "github.com/jwc20/sniffy/internal/scent"
)

const maxDepth = 10

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}

	s, _ := scent.Load(".")

	if s != nil && len(s.WatchPaths) > 0 {
		dirs = s.WatchPaths
	}

	m := model{
		dirs:    dirs,
		tests:   initTests(dirs, s),
		results: make(chan testResultMsg, 32),
		changes: make(chan fileChangedMsg, 32),
		scent:   s,
		styles:  DefaultStyles(),
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
