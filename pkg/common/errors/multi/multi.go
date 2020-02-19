/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package multi is an error type that holds multiple errors. These errors
// typically originate from operations that target multiple nodes.
// For example, a transaction proposal with two endorsers could return
// a multi error type if both endorsers return errors
package multi

import (
	"fmt"
	"strings"
)

// Errors is used to represent multiple errors
type Errors []error

// New Errors object with the given errors. Only non-nil errors are added.
func New(errs ...error) error {
	errors := Errors{}
	for _, err := range errs {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	return errors
}

// Append error to Errors. If the first arg is not an Errors object, one will be created
func Append(errs error, err error) error {
	m, ok := errs.(Errors)
	if !ok {
		return New(errs, err)
	}
	if err == nil {
		return errs
	}
	return append(m, err)
}

// ToError converts Errors to the error interface
// returns nil if no errors are present, a single error object if only one is present
func (errs Errors) ToError() error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errs
}

// Error implements the error interface to return a string representation of Errors
func (errs Errors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}

	errors := []string{fmt.Sprint("Multiple errors occurred:")}
	for _, err := range errs {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, " - ")
}
