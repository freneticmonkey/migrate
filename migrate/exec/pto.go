package exec

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/migrate/util/shell"
)

func executePTO(statement string, dryrun bool) (output string, err error) {

	params := []string{
		fmt.Sprintf("D=%s", "test"),
		fmt.Sprintf("t=%s", "test"),
		fmt.Sprintf("--alter \"%s\"", statement),
		"--critical-load \"Threads_running=500\"",
		"--execute",
	}

	if dryrun {
		output = fmt.Sprintf("PTO: [pt-online-schema-change %s]", strings.Join(params, " "))
	} else {
		output, err = shell.Run("pt-online-schema-change", "pto: ", params)
	}

	return output, err
}
