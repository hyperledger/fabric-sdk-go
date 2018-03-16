/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import (
	"errors"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
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
func ParseLevel(level string) (api.Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return api.Level(i), nil
		}
	}
	return api.ERROR, errors.New("logger: invalid log level")
}

//ParseString returns String repressentation of given log level
func ParseString(level api.Level) string {
	return levelNames[level]
}
