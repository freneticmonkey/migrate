package util

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

func LogInfo(info ...interface{}) {
	color.Set(color.FgWhite)
	log.Printf("INFO: %s", fmt.Sprintln(info...))
	color.Unset()
}

func LogInfof(format string, info ...interface{}) {
	if verbose {
		color.Set(color.FgWhite)
		log.Printf("INFO: %s", fmt.Sprintf(format, info...))
		color.Unset()
	}
}

func LogAttention(info ...interface{}) {
	if verbose {
		color.Set(color.FgYellow)
		log.Printf("INFO: %s", fmt.Sprintln(info...))
		color.Unset()
	}
}

func LogAttentionf(format string, info ...interface{}) {
	if verbose {
		color.Set(color.FgYellow)
		log.Printf("INFO: %s", fmt.Sprintf(format, info...))
		color.Unset()
	}
}

func LogWarn(warn ...interface{}) {
	if verbose {
		color.Set(color.FgMagenta)
		log.Printf("WARN: %s", fmt.Sprintln(warn...))
		color.Unset()
	}
}

func LogWarnf(format string, warn ...interface{}) {
	if verbose {
		color.Set(color.FgMagenta)
		log.Printf("WARN: %s", fmt.Sprintf(format, warn...))
		color.Unset()
	}
}

func LogAlert(alert ...interface{}) {
	if verbose {
		color.Set(color.FgCyan)
		log.Printf("ALERT: %s", fmt.Sprintln(alert...))
		color.Unset()
	}
}

func LogAlertf(format string, alert ...interface{}) {
	if verbose {
		color.Set(color.FgCyan)
		log.Printf("ALERT: %s", fmt.Sprintf(format, alert...))
		color.Unset()
	}
}

func LogError(err ...interface{}) {
	color.Set(color.FgHiRed)
	log.Printf("ERROR: %s", fmt.Sprintln(err...))
	color.Unset()
}

func LogErrorf(format string, err ...interface{}) {
	color.Set(color.FgHiRed)
	log.Printf("ERROR: %s", fmt.Sprintf(format, err...))
	color.Unset()
}

func LogFatal(code int, err ...interface{}) {
	LogErrorf("FATAL: %s", fmt.Sprintln(err...))
	os.Exit(code)
}
