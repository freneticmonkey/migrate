package util

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

func DebugDump(obj interface{}) {
	color.Set(color.FgCyan)
	spew.Dump(obj)
	color.Unset()
}
