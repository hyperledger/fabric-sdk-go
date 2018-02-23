/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package loglevel

// TODO: Leveler allows log levels to be enabled or disabled
//type Leveler interface {
//	SetLevel(module string, level Level)
//	GetLevel(module string) Level
//	IsEnabledFor(module string, level Level) bool
//	LogLevel(level string) (Level, error)
//}

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

//ModuleLevels maintains log levels based on module
type ModuleLevels struct {
	levels map[string]Level
}

// GetLevel returns the log level for the given module.
func (l *ModuleLevels) GetLevel(module string) Level {
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
func (l *ModuleLevels) SetLevel(module string, level Level) {
	if l.levels == nil {
		l.levels = make(map[string]Level)
	}
	l.levels[module] = level
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *ModuleLevels) IsEnabledFor(module string, level Level) bool {
	return level <= l.GetLevel(module)
}
