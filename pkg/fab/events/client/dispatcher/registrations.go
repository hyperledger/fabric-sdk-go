/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

// ConnectionReg is a connection registration
type ConnectionReg struct {
	Eventch chan<- *ConnectionEvent
}
