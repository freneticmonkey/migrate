package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/freneticmonkey/migrate/migrate/util"
)

// Run executes a command with arguments.  Both stdout and stderr are captured
// from the command and logged using the prefix defined in the shellPrefix parameter.
// The command output and any error is also returned.  This function blocks until
// the shell command is complete.
func Run(command string, shellPrefix string, args []string) (out string, err error) {

	var cmdReader io.ReadCloser
	var errReader io.ReadCloser

	cmd := exec.Command(command, args...)

	cmdReader, err = cmd.StdoutPipe()

	if err != nil {
		util.LogFatal(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		defer wg.Done()
		for scanner.Scan() {
			text := fmt.Sprintf(shellPrefix+" %s", scanner.Text())
			util.LogInfof(text)
			out += text
		}
	}()

	errScanner := bufio.NewScanner(errReader)
	go func() {
		wg.Done()
		for errScanner.Scan() {
			text := fmt.Sprintf(shellPrefix+" %s", scanner.Text())
			util.LogInfof(text)
			out += text
		}
	}()

	err = cmd.Start()
	if err != nil {
		util.LogFatal(os.Stderr, fmt.Sprintf("Error starting Cmd: [%s]", command), err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		util.LogFatal(os.Stderr, fmt.Sprintf("Error waiting for Cmd: [%s]", command), err)
		os.Exit(1)
	}

	wg.Wait()

	return out, err
}
