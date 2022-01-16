package log

import (
	"fmt"
	"io"
	"os"
)

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"

	CompleteTag = "[COMPLETE]"
	FailTag     = "[FAIL]"
)

var (
	LogTarget io.Writer = os.Stdout
)

type Logger struct {
	logTarget io.Writer
}

func NewLogger(logTarget io.Writer) *Logger {
	return &Logger{
		logTarget: logTarget,
	}
}

func formatColor(str, color string) string {
	return fmt.Sprintf("%s%s%s", color, str, colorReset)
}

func red(str string) string {
	return formatColor(str, colorRed)
}

func yellow(str string) string {
	return formatColor(str, colorYellow)
}

func green(str string) string {
	return formatColor(str, colorGreen)
}

func (l *Logger) TitleF(format string, args ...interface{}) {
	fmt.Fprintf(l.logTarget, "\n")
	fmt.Fprintf(l.logTarget, yellow("[TEST] "+format)+"\n", args...)
	fmt.Fprintf(l.logTarget, "============================================================================\n")
}

func (l *Logger) LogF(format string, args ...interface{}) {
	fmt.Fprintf(l.logTarget, "[LOG] "+format+"\n", args...)
}

func (l *Logger) ErrorF(format string, args ...interface{}) {
	fmt.Fprintf(l.logTarget, red("[ERROR] "+format)+"\n", args...)
}

func (l *Logger) Pass() {
	fmt.Fprint(l.logTarget, green("[PASS]")+"\n")
}

func (l *Logger) Fail() {
	fmt.Fprint(l.logTarget, red(FailTag)+"\n")
}

func (l *Logger) Complete() {
	fmt.Fprint(l.logTarget, green(CompleteTag)+"\n")
}

func (l *Logger) WriteString(str string) (int, error) {
	return l.logTarget.Write([]byte(str))
}
