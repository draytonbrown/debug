package errors

import "errors"

var (
	CommandNotFound = errors.New("command not found")
	NoCommandsFound = errors.New("no commands found")
)
