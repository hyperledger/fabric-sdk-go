/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"

type callerInfoKey struct {
	module string
	level  api.Level
}

//CallerInfo maintains module-level based information to toggle caller info
type CallerInfo struct {
	showcaller map[callerInfoKey]bool
}

//ShowCallerInfo enables caller info for given module and level
func (l *CallerInfo) ShowCallerInfo(module string, level api.Level) {
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = true
}

//HideCallerInfo disables caller info for given module and level
func (l *CallerInfo) HideCallerInfo(module string, level api.Level) {
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = false
}

//IsCallerInfoEnabled returns if callerinfo enabled for given module and level
func (l *CallerInfo) IsCallerInfoEnabled(module string, level api.Level) bool {
	showcaller, exists := l.showcaller[callerInfoKey{module, level}]
	if !exists {
		//If no callerinfo setting exists, then look for default
		showcaller, exists = l.showcaller[callerInfoKey{"", level}]
		if !exists {
			return true
		}
	}
	return showcaller
}

//getDefaultCallerInfoSetting default setting for callerinfo
func (l *CallerInfo) getDefaultCallerInfoSetting() map[callerInfoKey]bool {
	return map[callerInfoKey]bool{
		{"", api.CRITICAL}: true,
		{"", api.ERROR}:    true,
		{"", api.WARNING}:  true,
		{"", api.INFO}:     true,
		{"", api.DEBUG}:    true,
	}
}
