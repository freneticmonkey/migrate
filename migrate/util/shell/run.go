package shell

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// Run executes a command with arguments.  Both stdout and stderr are captured
// from the command and logged using the prefix defined in the shellPrefix parameter.
// The command output and any error is also returned.  This function blocks until
// the shell command is complete.
func Run(command string, shellPrefix string, args []string) (out string, err error) {

	var cmdout []string
	var errout []string

	var cmdReader io.ReadCloser
	var errReader io.ReadCloser

	cmd := exec.Command(command, args...)

	cmdReader, err = cmd.StdoutPipe()
	if err != nil {
		util.LogFatal(1, "Error creating StdoutPipe for Cmd", err)
	}

	errReader, err = cmd.StderrPipe()

	if err != nil {
		util.LogFatal(1, "Error creating StdoutError for Cmd", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		defer wg.Done()
		for scanner.Scan() {
			text := scanner.Text()
			util.LogInfof(fmt.Sprintf(shellPrefix+" %s ", text))
			cmdout = append(cmdout, text)
		}
	}()

	errScanner := bufio.NewScanner(errReader)
	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			util.LogInfof(fmt.Sprintf(shellPrefix+" %s ", text))
			errout = append(errout, text)
		}
	}()

	err = cmd.Start()
	if err != nil {
		util.LogErrorf("Error starting for Cmd: [%s] Error: [%s]", command, strings.Join(errout, "\n"))
		util.LogFatal(1, err)
	}

	err = cmd.Wait()
	if err != nil {
		util.LogErrorf("Error waiting for Cmd: [%s] Error: [%s]", command, strings.Join(errout, "\n"))
		util.LogFatal(1, err)
	}

	wg.Wait()

	return strings.Join(cmdout, "\n"), err
}
