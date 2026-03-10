package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

const maxDepth = 10

func main() {
	initColors()

	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}

	m := model{
		dirs:    dirs,
		tests:   initTests(dirs),
		results: make(chan testResultMsg, 32),
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
