package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) && strings.HasSuffix(event.Name, ".go") {
					log.Println("Created file:", event.Name)
					cmd := exec.Command("go", "test", "-json", "./...")
					output, _ := cmd.Output()
					fmt.Println(string(output))

				}
				if event.Has(fsnotify.Rename) {
					log.Println("Renamed file:", event.Name)
				}
				if event.Has(fsnotify.Chmod) {
					log.Println("Changed permission file:", event.Name)
				}
				if event.Has(fsnotify.Remove) {
					log.Println("Removed file:", event.Name)
				}
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add("./")
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}
