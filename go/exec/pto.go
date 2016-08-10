package exec

import (
	"fmt"
	"strings"

	"github.com/freneticmonkey/migrate/go/util"
)

func executePTO(statement string, dryrun bool) (output string, err error) {

	shell := util.GetShell()
	shell.SetPrefix("pto")

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
		output, err = shell.Run("pt-online-schema-change", params...)
	}

	return output, err
}
