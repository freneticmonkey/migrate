package util

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var verbose bool
var originalVerb bool

func SetVerbose(v bool) {
	verbose = v
}

func VerboseOverrideSet(ovr bool) {
	originalVerb = verbose
	verbose = ovr
}

func VerboseOverrideRestore() {
	verbose = originalVerb
}

func LogColour(out string, attr color.Attribute) {
	if verbose {
		color.Set(attr)
		log.Printf(out)
		color.Unset()
	}
}

func LogGreen(out string) {
	LogColour(out, color.FgGreen)
}

func LogWhite(out string) {
	LogColour(out, color.FgWhite)
}

func LogMagenta(out string) {
	LogColour(out, color.FgMagenta)
}

func LogYellow(out string) {
	LogColour(out, color.FgYellow)
}

func LogCyan(out string) {
	LogColour(out, color.FgCyan)
}

func LogRed(out string) {
	LogColour(out, color.FgRed)
}

func LogRedBright(out string) {
	LogColour(out, color.FgHiRed)
}

func LogInfo(info ...interface{}) {
	LogWhite("INFO:  " + fmt.Sprintln(info...))
}

func LogInfof(format string, info ...interface{}) {
	LogWhite("INFO:  " + fmt.Sprintf(format, info...))
}

func LogOk(info ...interface{}) {
	LogGreen("OK:    " + fmt.Sprintln(info...))
}

func LogOkf(format string, info ...interface{}) {
	LogGreen("OK:    " + fmt.Sprintf(format, info...))
}

func LogAttention(info ...interface{}) {
	LogYellow("ATTN:  " + fmt.Sprintln(info...))
}

func LogAttentionf(format string, info ...interface{}) {
	LogYellow("ATTN:  " + fmt.Sprintf(format, info...))
}

func LogWarn(warn ...interface{}) {
	LogMagenta("WARN:  " + fmt.Sprintln(warn...))
}

func LogWarnf(format string, warn ...interface{}) {
	LogMagenta("WARN:  " + fmt.Sprintf(format, warn...))
}

func LogAlert(alert ...interface{}) {
	LogCyan("ALERT: " + fmt.Sprintln(alert...))
}

func LogAlertf(format string, alert ...interface{}) {
	LogCyan("ALERT: " + fmt.Sprintf(format, alert...))
}

func LogError(err ...interface{}) {
	orig := verbose
	verbose = true
	LogRed("ERROR: " + fmt.Sprintln(err...))
	verbose = orig
}

func LogErrorf(format string, err ...interface{}) {
	orig := verbose
	verbose = true
	LogRed("ERROR: " + fmt.Sprintf(format, err...))
	verbose = orig
}

func LogFatal(code int, err ...interface{}) {
	LogErrorf("FATAL: %s", fmt.Sprintln(err...))
	os.Exit(code)
}
