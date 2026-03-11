package filewatcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jwc20/sniffa/internal/log"
)

const maxDepth = 7
const maxIdleTime = time.Hour

var floodThreshold = 250 * time.Millisecond

type Event struct {
	PkgPath  string
	Filename string
}

func Watch(ctx context.Context, dirs []string, run func(Event) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	if err := loadPaths(watcher, dirs); err != nil {
		return err
	}

	timer := time.NewTimer(maxIdleTime)
	defer timer.Stop()

	h := &fsEventHandler{last: time.Now(), fn: run}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return fmt.Errorf("exceeded idle timeout while watching files")
		case event := <-watcher.Events:
			resetTimer(timer)
			if handleDirCreated(watcher, event) {
				continue
			}
			if err := h.handleEvent(event); err != nil {
				return fmt.Errorf("failed to run tests for %v: %v", event.Name, err)
			}
		case err := <-watcher.Errors:
			return fmt.Errorf("failed while watching files: %v", err)
		}
	}
}

func resetTimer(timer *time.Timer) {
	if !timer.Stop() {
		<-timer.C
	}
	timer.Reset(maxIdleTime)
}

func loadPaths(watcher *fsnotify.Watcher, dirs []string) error {
	for _, dir := range FindAllDirs(dirs, maxDepth, nil) {
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to watch directory %v: %w", dir, err)
		}
	}
	return nil
}

func FindAllDirs(dirs []string, maxDepth int, exts []string) []string {
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}
	if len(exts) == 0 {
		exts = []string{".go"}
	}
	var output []string
	for _, dir := range dirs {
		const recur = "/..."
		if strings.HasSuffix(dir, recur) {
			dir = strings.TrimSuffix(dir, recur)
			output = append(output, findSubDirs(dir, maxDepth, exts)...)
			continue
		}
		output = append(output, dir)
	}
	return output
}

func findSubDirs(rootDir string, maxDepth int, exts []string) []string {
	var output []string
	maxDepth += pathDepth(rootDir)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("failed to watch %v: %v", path, err)
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if pathDepth(path) > maxDepth || exclude(path) {
			return filepath.SkipDir
		}
		if hasMatchingFiles(path, exts) {
			output = append(output, path)
		}
		return nil
	}
	filepath.Walk(rootDir, walker)
	return output
}

func pathDepth(path string) int {
	return strings.Count(filepath.Clean(path), string(filepath.Separator))
}

func exclude(path string) bool {
	base := filepath.Base(path)
	return (strings.HasPrefix(base, ".") && len(base) > 1) || base == "vendor" || base == "testdata"
}

func hasMatchingFiles(path string, exts []string) bool {
	fh, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fh.Close()
	for {
		names, err := fh.Readdirnames(20)
		if err == io.EOF {
			return false
		}
		if err != nil {
			return false
		}
		for _, name := range names {
			for _, ext := range exts {
				if strings.HasSuffix(name, ext) {
					return true
				}
			}
		}
	}
}

func handleDirCreated(watcher *fsnotify.Watcher, event fsnotify.Event) bool {
	if event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}
	info, err := os.Stat(event.Name)
	if err != nil || !info.IsDir() {
		return false
	}
	if err := watcher.Add(event.Name); err != nil {
		log.Warnf("failed to watch new directory %v: %v", event.Name, err)
	}
	return true
}

type fsEventHandler struct {
	last time.Time
	fn   func(Event) error
}

func (h *fsEventHandler) handleEvent(event fsnotify.Event) error {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
		return nil
	}
	if strings.HasPrefix(filepath.Base(event.Name), ".") {
		return nil
	}
	if time.Since(h.last) < floodThreshold {
		return nil
	}
	if err := h.fn(Event{
		PkgPath:  "./" + filepath.Dir(event.Name),
		Filename: event.Name,
	}); err != nil {
		return err
	}
	h.last = time.Now()
	return nil
}
