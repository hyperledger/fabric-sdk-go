/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deflogger

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

type moduleLeveled struct {
	levels map[string]apilogging.Level
}

// GetLevel returns the log level for the given module.
func (l *moduleLeveled) GetLevel(module string) apilogging.Level {
	level, exists := l.levels[module]
	if exists == false {
		level, exists = l.levels[""]
		// no configuration exists, default to info
		if exists == false {
			level = apilogging.INFO
		}
	}
	return level
}

// SetLevel sets the log level for the given module.
func (l *moduleLeveled) SetLevel(module string, level apilogging.Level) {
	if l.levels == nil {
		l.levels = make(map[string]apilogging.Level)
	}
	l.levels[module] = level
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *moduleLeveled) IsEnabledFor(module string, level apilogging.Level) bool {
	return level <= l.GetLevel(module)
}
