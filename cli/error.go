package cli

import "errors"

var ErrMissingInput = errors.New("no input provided")

type exitError struct {
	err     error
	code    int
	details string
}
