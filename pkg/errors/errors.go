/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package errors exposes the interface of github.com/hyperledger/fabric-sdk-go/pkg/errors
// and allows for SDK customizations. The initial implementation is a simple wrapper but we
// envision adding structured context and error codes.
package errors

import (
	perrors "github.com/pkg/errors"
)

// New calls pkg/errors New
func New(message string) error {
	return perrors.New(message)
}

// Errorf calls pkg/errors Errorf
func Errorf(format string, args ...interface{}) error {
	return perrors.Errorf(format, args...)
}

// WithStack calls pkg/errors WithStack
func WithStack(err error) error {
	return perrors.WithStack(err)
}

// Wrap calls pkg/errors Wrap
func Wrap(err error, message string) error {
	return perrors.Wrap(err, message)
}

// Wrapf calls pkg/errors Wrapf
func Wrapf(err error, format string, args ...interface{}) error {
	return perrors.Wrapf(err, format, args...)
}

// WithMessage calls pkg/errors WithMessage
func WithMessage(err error, message string) error {
	return perrors.WithMessage(err, message)
}

// Cause calls pkg/errors Cause
func Cause(err error) error {
	return perrors.Cause(err)
}
