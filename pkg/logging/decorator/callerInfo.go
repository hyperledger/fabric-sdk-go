/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package decorator

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

type callerInfoKey struct {
	module string
	level  apilogging.Level
}

//CallerInfo maintains module-level based information to toggle caller info
type CallerInfo struct {
	showcaller map[callerInfoKey]bool
}

//ShowCallerInfo enables caller info for given module and level
func (l *CallerInfo) ShowCallerInfo(module string, level apilogging.Level) {
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = true
}

//HideCallerInfo disables caller info for given module and level
func (l *CallerInfo) HideCallerInfo(module string, level apilogging.Level) {
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = false
}

//IsCallerInfoEnabled returns if callerinfo enabled for given module and level
func (l *CallerInfo) IsCallerInfoEnabled(module string, level apilogging.Level) bool {
	showcaller, exists := l.showcaller[callerInfoKey{module, level}]
	if exists == false {
		//If no callerinfo setting exists, then look for default
		showcaller, exists = l.showcaller[callerInfoKey{"", level}]
		if exists == false {
			return true
		}
	}
	return showcaller
}

//getDefaultCallerInfoSetting default setting for callerinfo
func (l *CallerInfo) getDefaultCallerInfoSetting() map[callerInfoKey]bool {
	return map[callerInfoKey]bool{
		callerInfoKey{"", apilogging.CRITICAL}: true,
		callerInfoKey{"", apilogging.ERROR}:    true,
		callerInfoKey{"", apilogging.WARNING}:  true,
		callerInfoKey{"", apilogging.INFO}:     true,
		callerInfoKey{"", apilogging.DEBUG}:    true,
	}
}
