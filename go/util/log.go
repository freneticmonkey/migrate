package util

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/robertkowalski/graylog-golang"
)

var verbose bool
var originalVerb bool
var filename string
var gelfDriver *gelf.Gelf
var gelfMessageFormat string

func SetLogFile(file string) func() {
	filename = file
	if file != "" {
		logFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Panicln(err)
		}

		log.SetOutput(io.MultiWriter(os.Stderr, logFile))

		return func() {
			e := logFile.Close()
			if e != nil {
				fmt.Fprintf(os.Stderr, "Problem closing the log file: %s\n", e)
			}
		}
	}
	return func() {
		LogError("Logging: Invalid Filename. Defaulting to stdout")
	}
}

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

	// If gelf is configured log regardless of verbosity
	if gelfDriver != nil {
		gelfDriver.Log(
			fmt.Sprintf(
				gelfMessageFormat,
				out,
			),
		)
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
