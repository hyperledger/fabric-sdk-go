/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"errors"
	"strings"
	"sync"
)

// ErrInvalidLogLevel is used when an invalid log level has been used.
var ErrInvalidLogLevel = errors.New("logger: invalid log level")

// Level defines all available log levels for log messages.
type Level int

// Log levels.
const (
	CRITICAL Level = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"INFO",
	"DEBUG",
}

// String returns the string representation of a logging level.
func (p Level) String() string {
	return levelNames[p]
}

// LogLevel returns the log level from a string representation.
func LogLevel(level string) (Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return Level(i), nil
		}
	}
	return ERROR, ErrInvalidLogLevel
}

// Leveled interface is the interface required to be able to add leveled
// logging.
type Leveled interface {
	GetLevel(string) Level
	SetLevel(Level, string)
	IsEnabledFor(Level, string) bool
}

type moduleLeveled struct {
	sync.RWMutex
	levels map[string]Level
}

// GetLevel returns the log level for the given module.
func (l *moduleLeveled) GetLevel(module string) Level {
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
func (l *moduleLeveled) SetLevel(level Level, module string) {
	l.Lock()
	defer l.Unlock()
	if l.levels == nil {
		l.levels = make(map[string]Level)
	}
	l.levels[module] = level
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *moduleLeveled) IsEnabledFor(level Level, module string) bool {
	return level <= l.GetLevel(module)
}
