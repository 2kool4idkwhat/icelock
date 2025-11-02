package log

import (
	"log"
	"os"
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var (
	error = log.New(os.Stderr, "icelock ERROR: ", 0)
	warn  = log.New(os.Stderr, "icelock WARN: ", 0)
	info  = log.New(os.Stderr, "icelock INFO: ", 0)
	debug = log.New(os.Stderr, "icelock DEBUG: ", 0)

	currentLevel = LevelInfo // default level
)

// SetLevel sets the logging level
func SetLevel(level string) {
	switch level {
	case "error":
		currentLevel = LevelError
	case "warn":
		currentLevel = LevelWarn
	case "info":
		currentLevel = LevelInfo
	case "debug":
		currentLevel = LevelDebug
	default:
		currentLevel = LevelWarn
	}
}

// Error logs an error message
func Error(format string, v ...any) {

	if currentLevel >= LevelError {
		error.Printf(format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...any) {

	if currentLevel >= LevelWarn {
		warn.Printf(format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...any) {

	if currentLevel >= LevelInfo {
		info.Printf(format, v...)
	}
}

// Debug logs a debug message
func Debug(format string, v ...any) {

	if currentLevel >= LevelDebug {
		debug.Printf(format, v...)
	}
}
