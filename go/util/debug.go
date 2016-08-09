package util

import (
	"fmt"
	"strings"

	"github.com/aryann/difflib"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

// DebugDump Pretty print a data structure
func DebugDump(obj interface{}) {
	color.Set(color.FgCyan)
	spew.Dump(obj)
	color.Unset()
}

// DebugDumpDiff Pretty print differences between data structures
func DebugDumpDiff(left interface{}, right interface{}) {

	l := strings.Split(spew.Sdump(left), "\n")
	r := strings.Split(spew.Sdump(right), "\n")

	DebugDiffStrings(l, r)
}

//DebugDiffString Diff two strings
func DebugDiffString(l, r string) {
	// If the string contains a newline, then split on newlines first
	if strings.Contains(l, "\n") {
		DebugDiffStrings(strings.Split(l, "\n"), strings.Split(r, "\n"))
	} else {
		DebugDiffStrings([]string{l}, []string{r})
	}
}

// DebugDiffStrings Diff two arrays of strings
func DebugDiffStrings(l, r []string) {

	diffs := difflib.Diff(l, r)

	for _, diff := range diffs {
		var prefix string
		switch diff.Delta {
		case difflib.Common:
			prefix = "    "
			color.Set(color.FgGreen)
		case difflib.LeftOnly:
			prefix = " << "
			color.Set(color.FgRed)
		case difflib.RightOnly:
			prefix = " >> "
			color.Set(color.FgYellow)
		}
		fmtStr := fmt.Sprintf("%s %s", prefix, diff.Payload)
		fmt.Println(fmtStr)
	}
	color.Unset()
}
