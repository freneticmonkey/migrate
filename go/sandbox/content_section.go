package sandbox

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
)

const (
	preserveBegin = "PRESERVE_BEGIN"
	preserveEnd   = "PRESERVE_END"
)

type contentMap map[string]string

var content contentMap

func contentSlice(contentName string) (contents string) {
	if contents, ok := content[contentName]; ok {
		return contents
	}
	return ""
}

func parseContent(fileContents string) error {
	content = make(contentMap)

	exp := fmt.Sprintf("%s|%s", preserveBegin, preserveEnd)
	pattern, err := regexp.Compile(exp)
	if err != nil {
		util.LogErrorf("Couldn't compile: %v", err)
		return err
	}

	// Extract each section
	sections := pattern.Split(fileContents, -1)

	for i := 1; i < len(sections); i += 2 {
		section := sections[i]
		// Extract the name
		sectionLines := strings.Split(section, "\n")
		name := strings.Trim(sectionLines[0], "[] ")
		content[name] = strings.TrimSpace(strings.Join(sectionLines[1:len(sectionLines)-1], "\n"))

	}

	return nil
}
