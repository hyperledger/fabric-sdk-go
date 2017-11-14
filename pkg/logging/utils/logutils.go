/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"errors"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
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
func LogLevel(level string) (apilogging.Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return apilogging.Level(i), nil
		}
	}
	return apilogging.ERROR, errors.New("logger: invalid log level")
}

//LogLevelString returns String repressentation of given log level
func LogLevelString(level apilogging.Level) string {
	return levelNames[level]
}
