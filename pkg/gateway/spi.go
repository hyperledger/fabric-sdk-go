/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// This contains the service provider interface (SPI) which provides the mechanism
// for implementing alternative gateway strategies, wallets, etc.
// This is currently experimental and will be implemented in future user stories

// CommitHandlerFactory is currently unimplemented
type CommitHandlerFactory interface {
	Create(string, Network) CommitHandler
}

// CommitHandler is currently unimplemented
type CommitHandler interface {
	StartListening()
	WaitForEvents(int64)
	CancelListening()
}

// Identity is the base type for implementing wallet identities - experimental
type Identity struct {
	theType string
}

// IdentityType represents a specific identity format - experimental
type IdentityType interface {
	Type() string
}

// IDHandler represents the storage of identity information - experimental
type IDHandler interface {
	GetElements(id IdentityType) map[string]string
	FromElements(map[string]string) IdentityType
}
