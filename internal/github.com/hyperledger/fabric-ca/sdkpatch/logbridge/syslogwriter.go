/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package logbridge

// cLogger implements CFSSL's SyslogWriter interface
type cLogger struct {
}

// Debug bridges calls to the Go SDK logger's Debug.
func (log *cLogger) Debug(s string) {
	logger.Debug(s)
}

// Info bridges calls to the Go SDK logger's Info.
func (log *cLogger) Info(s string) {
	logger.Info(s)
}

// Warning bridges calls to the Go SDK logger's Warn.
func (log *cLogger) Warning(s string) {
	logger.Warn(s)
}

// Err bridges calls to the Go SDK logger's Error.
func (log *cLogger) Err(s string) {
	logger.Error(s)
}

// Crit bridges calls to the Go SDK logger's Error.
func (log *cLogger) Crit(s string) {
	logger.Error(s)
}

// Emerg bridges calls to the Go SDK logger's Error.
func (log *cLogger) Emerg(s string) {
	logger.Error(s)
}
