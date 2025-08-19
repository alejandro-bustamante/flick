package transformations

import "fmt"

type Logger struct {
	level string
}

func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.level == "debug" {
		fmt.Printf("[DEBUG] "+msg+"\n", args...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if l.level == "debug" || l.level == "info" {
		fmt.Printf("[INFO] "+msg+"\n", args...)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}
