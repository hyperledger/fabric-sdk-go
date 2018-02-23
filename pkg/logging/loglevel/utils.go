/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package loglevel

import (
	"errors"
	"strings"
)

//Log level names in string
var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"INFO",
	"DEBUG",
}

// ParseLevel returns the log level from a string representation.
func ParseLevel(level string) (Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return Level(i), nil
		}
	}
	return ERROR, errors.New("logger: invalid log level")
}

//ParseString returns String repressentation of given log level
func ParseString(level Level) string {
	return levelNames[level]
}
