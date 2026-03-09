package main

import (
	"os"
	"log"

	sniffa "github.com/jwc20/sniffa"
)

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./..."}
	}
	if err := sniffa.RunWatcher(dirs, true); err != nil {
		log.Fatal(err)
	}
}
