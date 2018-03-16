/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
)

// RegisterConnectionEvent is a request to register for connection events
type RegisterConnectionEvent struct {
	esdispatcher.RegisterEvent
	Reg *ConnectionReg
}

// NewRegisterConnectionEvent creates a new RegisterConnectionEvent
func NewRegisterConnectionEvent(eventch chan<- *ConnectionEvent, regch chan<- fab.Registration, errch chan<- error) *RegisterConnectionEvent {
	return &RegisterConnectionEvent{
		Reg:           &ConnectionReg{Eventch: eventch},
		RegisterEvent: esdispatcher.NewRegisterEvent(regch, errch),
	}
}

// ConnectedEvent indicates that the client has connected to the server
type ConnectedEvent struct {
}

// NewConnectedEvent creates a new ConnectedEvent
func NewConnectedEvent() *ConnectedEvent {
	return &ConnectedEvent{}
}

// DisconnectedEvent indicates that the client has disconnected from the server
type DisconnectedEvent struct {
	Err error
}

// NewDisconnectedEvent creates a new DisconnectedEvent
func NewDisconnectedEvent(err error) *DisconnectedEvent {
	return &DisconnectedEvent{Err: err}
}

// ConnectEvent is a request to connect to the server
type ConnectEvent struct {
	ErrCh        chan<- error
	FromBlockNum uint64
}

// NewConnectEvent creates a new ConnectEvent
func NewConnectEvent(errch chan<- error) *ConnectEvent {
	return &ConnectEvent{ErrCh: errch}
}

// DisconnectEvent is a request to disconnect to the server
type DisconnectEvent struct {
	Errch chan<- error
}

// NewDisconnectEvent creates a new DisconnectEvent
func NewDisconnectEvent(errch chan<- error) *DisconnectEvent {
	return &DisconnectEvent{Errch: errch}
}

// ConnectionEvent is sent when the client disconnects from or
// reconnects to the event server. Connected == true means that the
// client has connected, whereas Connected == false means that the
// client has disconnected. In the disconnected case, Err contains
// the disconnect error.
type ConnectionEvent struct {
	Connected bool
	Err       error
}

// NewConnectionEvent returns a new ConnectionEvent
func NewConnectionEvent(connected bool, err error) *ConnectionEvent {
	return &ConnectionEvent{Connected: connected, Err: err}
}
