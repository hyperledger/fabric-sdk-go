/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import "sync"

type callerInfo struct {
	sync.RWMutex
	showcaller map[string]bool
}

func (l *callerInfo) ShowCallerInfo(module string) {
	l.RLock()
	defer l.RUnlock()
	if l.showcaller == nil {
		l.showcaller = make(map[string]bool)
	}
	l.showcaller[module] = true
}

func (l *callerInfo) HideCallerInfo(module string) {
	l.Lock()
	defer l.Unlock()
	if l.showcaller == nil {
		l.showcaller = make(map[string]bool)
	}
	l.showcaller[module] = false
}

func (l *callerInfo) IsCallerInfoEnabled(module string) bool {
	showcaller, exists := l.showcaller[module]
	if exists == false {
		showcaller, exists = l.showcaller[""]
		// no configuration exists, default to false
		if exists == false {
			return false
		}
	}
	return showcaller
}
