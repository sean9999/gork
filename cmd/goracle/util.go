package main

import (
	"fmt"
	"io"
	"os"
)

// an CLIError is an error with an exit code
type CLIError struct {
	Msg      string
	ExitCode int
	Child    error
}

func (o *CLIError) Error() string {
	if o.Child == nil {
		return o.Msg
	} else {
		return fmt.Sprintf("%s: %s", o.Msg, o.Child)
	}
}

func (o *CLIError) Wrap(child error) {
	o.Child = child
}

func (o *CLIError) Unrawp() error {
	return o.Child
}

func complain(msg string, exitCode int, child error, stream io.Writer) {
	err := &CLIError{msg, exitCode, child}
	fmt.Fprintln(stream, err)
	os.Exit(exitCode)
}
