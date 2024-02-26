package log

import (
	"fmt"
	"os"
)

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"

	CompleteTag = "[COMPLETE]"
	FailTag     = "[FAIL]"
	WarningTag  = "[WARNING]"
	SkipTag     = "[SKIP]"
	PassTag     = "[PASS]"
	ErrorTag    = "[ERROR]"
	TestTag     = "[TEST]"
	LogTag      = "[LOG]"
)

type Logger struct {
	file *os.File
}

func NewLogger(logFile *os.File) *Logger {
	return &Logger{
		file: logFile,
	}
}

func (l *Logger) FileName() string {
	return l.file.Name()
}

func (l *Logger) WriteStringF(format string, args ...interface{}) (int, error) {
	format += "\n"
	fmt.Printf(format, args...)

	if l.file == nil {
		return 0, nil
	}

	return l.file.Write([]byte(fmt.Sprintf(format, args...)))
}

func formatColor(str, color string) string {
	return fmt.Sprintf("%s%s%s", color, str, colorReset)
}

func Red(str string) string {
	return formatColor(str, colorRed)
}

func blue(str string) string {
	return formatColor(str, colorBlue)
}

func Green(str string) string {
	return formatColor(str, colorGreen)
}

func yellow(str string) string {
	return formatColor(str, colorYellow)
}

func (l *Logger) TitleF(format string, args ...interface{}) {
	l.WriteStringF("")
	l.WriteStringF(blue(TestTag+" "+format), args...)
	l.WriteStringF("--------------------------------------------------")
}

func (l *Logger) LogF(format string, args ...interface{}) {
	l.WriteStringF(LogTag+" "+format, args...)
}

func (l *Logger) ErrorF(format string, args ...interface{}) {
	l.WriteStringF(Red(ErrorTag+" "+format), args...)
}

func (l *Logger) WarningF(format string, args ...interface{}) {
	l.WriteStringF(yellow(WarningTag+" "+format), args...)
}

func (l *Logger) Warning() {
	l.WriteStringF(yellow(WarningTag))
}

func (l *Logger) Pass() {
	l.WriteStringF(Green(PassTag))
}

func (l *Logger) Skip() {
	l.WriteStringF(yellow(SkipTag))
}

func (l *Logger) Fail() {
	l.WriteStringF(Red(FailTag))
}

func (l *Logger) Complete() {
	l.WriteStringF(Green(CompleteTag))
}
