/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"errors"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"
)

//Log level names in string
var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"INFO",
	"DEBUG",
}

// LogLevel returns the log level from a string representation.
func LogLevel(level string) (api.Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return api.Level(i), nil
		}
	}
	return api.ERROR, errors.New("logger: invalid log level")
}

//LogLevelString returns String repressentation of given log level
func LogLevelString(level api.Level) string {
	return levelNames[level]
}
