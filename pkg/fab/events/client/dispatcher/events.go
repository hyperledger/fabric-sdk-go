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

// DisconnectedError is the error that is associated with the disconnect.
type DisconnectedError interface {
	error

	// IsFatal returns true if the error is fatal, meaning that a reconnect attempt would not succeed
	IsFatal() bool
}

type disconnectedError struct {
	cause error
	fatal bool
}

func (e *disconnectedError) Error() string {
	return e.cause.Error()
}

func (e *disconnectedError) IsFatal() bool {
	return e.fatal
}

// DisconnectedEvent indicates that the client has disconnected from the server
type DisconnectedEvent struct {
	Err DisconnectedError
}

// NewDisconnectedEvent creates a new DisconnectedEvent
func NewDisconnectedEvent(cause error) *DisconnectedEvent {
	return &DisconnectedEvent{Err: &disconnectedError{cause: cause}}
}

// NewFatalDisconnectedEvent creates a new DisconnectedEvent which indicates that a reconnect is not possible
func NewFatalDisconnectedEvent(cause error) *DisconnectedEvent {
	return &DisconnectedEvent{Err: &disconnectedError{cause: cause, fatal: true}}
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
	Err       DisconnectedError
}

// NewConnectionEvent returns a new ConnectionEvent
func NewConnectionEvent(connected bool, err DisconnectedError) *ConnectionEvent {
	return &ConnectionEvent{Connected: connected, Err: err}
}
