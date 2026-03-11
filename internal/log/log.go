package log

import (
    "fmt"
    "os"
)

type Level uint8

const (
    WarnLevel Level = iota
    DebugLevel
)

var level = WarnLevel

func SetLevel(l Level) {
    level = l
}

func Warnf(format string, args ...interface{}) {
    if level < WarnLevel {
        return
    }
    fmt.Fprintf(os.Stderr, "WARN "+format+"\n", args...)
}

func Debugf(format string, args ...interface{}) {
    if level < DebugLevel {
        return
    }
    fmt.Fprintf(os.Stderr, format+"\n", args...)
}
