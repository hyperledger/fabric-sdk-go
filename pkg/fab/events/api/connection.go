/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// Connection defines the functions for an event server connection
type Connection interface {
	// Receive sends events to the given channel
	Receive(chan<- interface{})
	// Close closes the connection
	Close()
	// Closed return true if the connection is closed
	Closed() bool
}

// ConnectionProvider creates a Connection.
type ConnectionProvider func(channelID string, context context.Client, peer fab.Peer) (Connection, error)
