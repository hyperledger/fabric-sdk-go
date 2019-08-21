/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// BlockEvent contains the data for the block event
type BlockEvent struct {
	// Block is the block that was committed
	Block *cb.Block
	// SourceURL specifies the URL of the peer that produced the event
	SourceURL string
}

// FilteredBlockEvent contains the data for a filtered block event
type FilteredBlockEvent struct {
	// FilteredBlock contains a filtered version of the block that was committed
	FilteredBlock *pb.FilteredBlock
	// SourceURL specifies the URL of the peer that produced the event
	SourceURL string
}

// TxStatusEvent contains the data for a transaction status event
type TxStatusEvent struct {
	// TxID is the ID of the transaction in which the event was set
	TxID string
	// TxValidationCode is the status code of the commit
	TxValidationCode pb.TxValidationCode
	// BlockNumber contains the block number in which the
	// transaction was committed
	BlockNumber uint64
	// SourceURL specifies the URL of the peer that produced the event
	SourceURL string
}

// CCEvent contains the data for a chaincode event
type CCEvent struct {
	// TxID is the ID of the transaction in which the event was set
	TxID string
	// ChaincodeID is the ID of the chaincode that set the event
	ChaincodeID string
	// EventName is the name of the chaincode event
	EventName string
	// Payload contains the payload of the chaincode event
	// NOTE: Payload will be nil for filtered events
	Payload []byte
	// BlockNumber contains the block number in which the
	// chaincode event was committed
	BlockNumber uint64
	// SourceURL specifies the URL of the peer that produced the event
	SourceURL string
}

// Registration is a handle that is returned from a successful RegisterXXXEvent.
// This handle should be used in Unregister in order to unregister the event.
type Registration interface{}

// BlockFilter is a function that determines whether a Block event
// should be ignored
type BlockFilter func(block *cb.Block) bool

// EventService is a service that receives events such as block, filtered block,
// chaincode, and transaction status events.
type EventService interface {
	// RegisterBlockEvent registers for block events. If the caller does not have permission
	// to register for block events then an error is returned.
	// Note that Unregister must be called when the registration is no longer needed.
	// - filter is an optional filter that filters out unwanted events. (Note: Only one filter may be specified.)
	// - Returns the registration and a channel that is used to receive events. The channel
	//   is closed when Unregister is called.
	RegisterBlockEvent(filter ...BlockFilter) (Registration, <-chan *BlockEvent, error)

	// RegisterFilteredBlockEvent registers for filtered block events.
	// Note that Unregister must be called when the registration is no longer needed.
	// - Returns the registration and a channel that is used to receive events. The channel
	//   is closed when Unregister is called.
	RegisterFilteredBlockEvent() (Registration, <-chan *FilteredBlockEvent, error)

	// RegisterChaincodeEvent registers for chaincode events.
	// Note that Unregister must be called when the registration is no longer needed.
	// - ccID is the chaincode ID for which events are to be received
	// - eventFilter is the chaincode event filter (regular expression) for which events are to be received
	// - Returns the registration and a channel that is used to receive events. The channel
	//   is closed when Unregister is called.
	RegisterChaincodeEvent(ccID, eventFilter string) (Registration, <-chan *CCEvent, error)

	// RegisterTxStatusEvent registers for transaction status events.
	// Note that Unregister must be called when the registration is no longer needed.
	// - txID is the transaction ID for which events are to be received
	// - Returns the registration and a channel that is used to receive events. The channel
	//   is closed when Unregister is called.
	RegisterTxStatusEvent(txID string) (Registration, <-chan *TxStatusEvent, error)

	// Unregister removes the given registration and closes the event channel.
	// - reg is the registration handle that was returned from one of the Register functions
	Unregister(reg Registration)
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

// EventSnapshot contains a snapshot of the event client before it was stopped.
// The snapshot includes all of the event registrations and the last block received.
type EventSnapshot interface {
	// LastBlockReceived returns the block number of the last block received at the time
	// that the snapshot was taken.
	LastBlockReceived() uint64

	// BlockRegistrations returns the block registrations.
	BlockRegistrations() []Registration

	// FilteredBlockRegistrations returns the filtered block registrations.
	FilteredBlockRegistrations() []Registration

	// CCRegistrations returns the chaincode registrations.
	CCRegistrations() []Registration

	// TxStatusRegistrations returns the transaction status registrations.
	TxStatusRegistrations() []Registration

	// Closes all registrations
	Close()
}

// EventClient is a client that connects to a peer and receives channel events
// such as block, filtered block, chaincode, and transaction status events.
type EventClient interface {
	EventService

	// Connect connects to the event server.
	Connect() error

	// Close closes the connection to the event server and releases all resources.
	// Once this function is invoked the client may no longer be used.
	Close()

	// CloseIfIdle closes the connection to the event server only if there are no outstanding
	// registrations.
	// Returns true if the client was closed. In this case the client may no longer be used.
	// A return value of false indicates that the client could not be closed since
	// there was at least one registration.
	CloseIfIdle() bool

	// TransferRegistrations transfers all registrations into an EventSnapshot.
	// The registrations are not closed and may be transferred to a new event client.
	// - close: If true then the client will also be closed
	TransferRegistrations(close bool) (EventSnapshot, error)
}
