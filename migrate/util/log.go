package util

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

func LogInfo(info ...interface{}) {
	color.Set(color.FgWhite)
	log.Printf("INFO: %s", fmt.Sprintln(info...))
	color.Unset()
}

func LogInfof(format string, info ...interface{}) {
	color.Set(color.FgWhite)
	log.Printf("INFO: %s", fmt.Sprintf(format, info...))
	color.Unset()
}

func LogWarn(warn ...interface{}) {
	color.Set(color.FgMagenta)
	log.Printf("WARN: %s", fmt.Sprintln(warn...))
	color.Unset()
}

func LogWarnf(format string, warn ...interface{}) {
	color.Set(color.FgMagenta)
	log.Printf("WARN: %s", fmt.Sprintf(format, warn...))
	color.Unset()
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

func LogFatal(err ...interface{}) {
	color.Set(color.FgHiRed)
	log.Fatalf("FATAL: %s", fmt.Sprintln(err...))
}
