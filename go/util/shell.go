package util

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// ShellRunner An interface to abstract shell command structs
type ShellRunner interface {
	SetPrefix(string)
	Run(string, ...string) (string, error)
}

// ShellExecutor Shell Command wrapper
type ShellExecutor struct {
	prefix string
}

// SetPrefix Sets the prefix that will be used for console output
func (e *ShellExecutor) SetPrefix(prefix string) {
	e.prefix = prefix
}

// Run executes a command with arguments.  Both stdout and stderr are captured
// from the command and logged using the prefix defined in the shellPrefix parameter.
// The command output and any error is also returned.  This function blocks until
// the shell command is complete.
func (e ShellExecutor) Run(command string, args ...string) (out string, err error) {

	var cmdout []string
	var errout []string

	var cmdReader io.ReadCloser
	var errReader io.ReadCloser

	cmd := exec.Command(command, args...)

	cmdReader, err = cmd.StdoutPipe()
	if err != nil {
		LogFatal(1, "Error creating StdoutPipe for Cmd", err)
	}

	errReader, err = cmd.StderrPipe()

	if err != nil {
		LogFatal(1, "Error creating StdoutError for Cmd", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		defer wg.Done()
		for scanner.Scan() {
			text := scanner.Text()
			if e.prefix != "" {
				LogInfo(e.prefix + ": " + text)
			} else {
				LogInfo(text)
			}
			cmdout = append(cmdout, text)
		}
	}()

	errScanner := bufio.NewScanner(errReader)
	go func() {
		defer wg.Done()
		for errScanner.Scan() {
			text := errScanner.Text()
			if e.prefix != "" {
				LogInfo(e.prefix + ": " + text)
			} else {
				LogInfo(text)
			}
			errout = append(errout, text)
		}
	}()

	err = cmd.Start()
	if err != nil {
		LogErrorf("Error starting for Cmd: [%s] Error: [%s]", command, strings.Join(errout, "\n"))
		LogFatal(1, err)
	}

	err = cmd.Wait()
	if err != nil {
		LogErrorf("Error waiting for Cmd: [%s] Error: [%s]", command, strings.Join(errout, "\n"))
		LogFatal(1, err)
	}

	wg.Wait()

	return strings.Join(cmdout, "\n"), err

}

// Lambda Callback type to allow simulated execution of shell commands
type Lambda func(string, []string) error

// ExpectedCommand Defines a mock command input and output
type ExpectedCommand struct {
	Cmd       string
	Args      []string
	Triggered bool
	result    string
	err       error
	lambda    Lambda
}

// String String representation of the mock command
func (c ExpectedCommand) String() string {
	return fmt.Sprintf("%s %s", c.Cmd, strings.Join(c.Args, " "))
}

// Compare Check if the input command matches this mock command
func (c ExpectedCommand) Compare(cmd string, args ...string) error {
	if cmd != c.Cmd {
		return fmt.Errorf("ExpectedCommand Mismatch. Expected Command: [%s] Detected: [%s]", c.Cmd, cmd)
	}

	if len(c.Args) != len(args) {
		return fmt.Errorf("ExpectedCommand Mismatch. Command: [%s]: Expected [%d] args.  Received: [%d]", c.String(), len(c.Args), len(args))
	}

	for i, arg := range args {
		if arg != c.Args[i] {
			return fmt.Errorf("ExpectedCommand Mismatch. Command: [%s]: argument in position %d expected [%s].  Received: [%s]", c.String(), i, c.Args[i], arg)
		}
	}
	return nil
}

// GetResult Return the command result with the correct prefix
func (c ExpectedCommand) GetResult(prefix string) string {
	if len(prefix) > 0 {
		lines := strings.Split(c.result, "\n")
		joiner := fmt.Sprintf("\n%s: ", prefix)
		LogInfof("%s: %s", prefix, strings.Join(lines, joiner))
	}
	return c.result
}

// Lambda Execute simulated function
func (c ExpectedCommand) Lambda() error {
	if c.lambda != nil {
		return c.lambda(c.Cmd, c.Args)
	}
	return nil
}

// MockShellExecutor A mock command executor that fulfills the Runner interface
// modelled after DATA-DOG/go-sqlmock
type MockShellExecutor struct {
	prefix       string
	expectations []ExpectedCommand
	next         *ExpectedCommand
}

func (e MockShellExecutor) Count() int {
	return len(e.expectations)
}

// ExpectExec Create a expected command
func (e *MockShellExecutor) ExpectExec(cmd string, args []string, output string, err error) {
	e.expectations = append(e.expectations, ExpectedCommand{
		Cmd:    cmd,
		Args:   args,
		result: output,
		err:    err,
	})
}

// ExpectExecWithLambda Create a expected command
func (e *MockShellExecutor) ExpectExecWithLambda(cmd string, args []string, output string, err error, lb Lambda) {
	e.expectations = append(e.expectations, ExpectedCommand{
		Cmd:    cmd,
		Args:   args,
		result: output,
		err:    err,
		lambda: lb,
	})
}

// SetPrefix Set a prefix on the command output
func (e *MockShellExecutor) SetPrefix(prefix string) {
	e.prefix = prefix
}

// Run Execute a command.  This function checks all expected commands for a match in order of
// creation and returns the expected output or error if found.
func (e *MockShellExecutor) Run(cmd string, args ...string) (out string, err error) {
	var expected *ExpectedCommand
	var fulfilled int

	for i, next := range e.expectations {
		if next.Triggered {
			fulfilled++
			continue
		}
		expected = &e.expectations[i]
		break
	}
	if expected == nil {
		msg := "ExpectedCommand: [%s] with Args [%+v], was not expected"
		if fulfilled == len(e.expectations) {
			msg = " all expectations were already fulfilled, " + msg
		}
		return "", fmt.Errorf(msg, cmd, args)
	}

	if err = expected.Compare(cmd, args...); err != nil {
		return "", err
	}

	expected.Triggered = true
	LogAlert("Mock ExpectedCommand >> DETECTED: ###> [" + expected.String() + "]\n\n")

	// Trigger lambda
	err = expected.Lambda()
	if err != nil {
		return "", err
	}

	if expected.err != nil {
		return "", expected.err // mocked to return error
	}

	return expected.GetResult(e.prefix), err
}

func (e *MockShellExecutor) ExpectationsWereMet() error {
	for _, ex := range e.expectations {
		if !ex.Triggered {
			return fmt.Errorf("there is a remaining expectation which was not matched: %s", ex)
		}
	}
	return nil
}
