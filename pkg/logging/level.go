/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"strings"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// ErrInvalidLogLevel is used when an invalid log level has been used.
var ErrInvalidLogLevel = errors.New("logger: invalid log level")

// Log levels.
const (
	CRITICAL apilogging.Level = iota
	ERROR
	WARNING
	INFO
	DEBUG
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
	return ERROR, ErrInvalidLogLevel
}

type moduleLeveled struct {
	sync.RWMutex
	levels map[string]apilogging.Level
}

// GetLevel returns the log level for the given module.
func (l *moduleLeveled) GetLevel(module string) apilogging.Level {
	l.RLock()
	defer l.RUnlock()
	level, exists := l.levels[module]
	if exists == false {
		level, exists = l.levels[""]
		// no configuration exists, default to info
		if exists == false {
			level = INFO
		}
	}
	return level
}

// SetLevel sets the log level for the given module.
func (l *moduleLeveled) SetLevel(level apilogging.Level, module string) {
	l.Lock()
	defer l.Unlock()
	if l.levels == nil {
		l.levels = make(map[string]apilogging.Level)
	}
	l.levels[module] = level
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *moduleLeveled) IsEnabledFor(level apilogging.Level, module string) bool {
	return level <= l.GetLevel(module)
}
