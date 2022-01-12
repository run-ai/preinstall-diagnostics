package log

import (
	"fmt"
	"io"
	"os"
)

const CompleteTag = "[COMPLETE]"

var (
	LogTarget io.Writer = os.Stdout
)

func TitleF(format string, args ...interface{}) {
	fmt.Fprintf(LogTarget, "\n")
	fmt.Fprintf(LogTarget, "[TEST] "+format+"\n", args...)
	fmt.Fprintf(LogTarget, "============================================================================\n")
}

func LogF(format string, args ...interface{}) {
	fmt.Fprintf(LogTarget, "[LOG] "+format+"\n", args...)
}

func ErrorF(format string, args ...interface{}) {
	fmt.Fprintf(LogTarget, "[ERROR] "+format+"\n", args...)
}

func Pass() {
	fmt.Fprint(LogTarget, "[PASS]\n")
}

func Fail() {
	fmt.Fprint(LogTarget, "*****************[FAIL]*******************\n")
}

func Complete() {
	fmt.Fprint(LogTarget, CompleteTag+"\n")
}
