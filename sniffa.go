package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jwc20/sniffa/internal/filewatcher"
)

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}
	if err := runWatcher(dirs, true); err != nil {
		log.Fatal(err)
	}
}

func runWatcher(dirs []string, clearScreen bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return filewatcher.Watch(ctx, dirs, clearScreen, runTests)
}

func runTests(event filewatcher.Event) error {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v", event.PkgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tests failed: %v\n", err)
	}
	return nil
}
