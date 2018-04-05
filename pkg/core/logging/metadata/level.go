/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"

//ModuleLevels maintains log levels based on module
type ModuleLevels struct {
	levels map[string]api.Level
}

// GetLevel returns the log level for the given module.
func (l *ModuleLevels) GetLevel(module string) api.Level {
	level, exists := l.levels[module]
	if !exists {
		level, exists = l.levels[""]
		// no configuration exists, default to info
		if !exists {
			level = api.INFO
		}
	}
	return level
}

// SetLevel sets the log level for the given module.
func (l *ModuleLevels) SetLevel(module string, level api.Level) {
	if l.levels == nil {
		l.levels = make(map[string]api.Level)
	}
	l.levels[module] = level
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *ModuleLevels) IsEnabledFor(module string, level api.Level) bool {
	return level <= l.GetLevel(module)
}
