package utils

import (
	"errors"
	"strings"
)

type Errors []error

func NewErrors() Errors {
	errs := make(Errors, 0)
	return errs
}

func (errs Errors) Merged() error {
	if errs == nil || len(errs) == 0 {
		return nil
	}

	errStrings := make([]string, 0)
	for _, err := range errs {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return errors.New(strings.Join(errStrings, "; "))
}
