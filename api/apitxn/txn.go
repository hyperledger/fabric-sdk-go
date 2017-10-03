/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apitxn

import (
	"time"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// QueryRequest contains the parameters for query
type QueryRequest struct {
	ChaincodeID string
	Fcn         string
	Args        [][]byte
}

// QueryOpts allows the user to specify more advanced options
type QueryOpts struct {
	Notifier           chan QueryResponse  // async
	ProposalProcessors []ProposalProcessor // targets
	Timeout            time.Duration
}

// QueryResponse contains result of asynchronous call
type QueryResponse struct {
	Response []byte
	Error    error
}

// ExecuteTxResponse contains result of asynchronous call
type ExecuteTxResponse struct {
	Response         TransactionID
	Error            error
	TxValidationCode pb.TxValidationCode
}

// ExecuteTxRequest contains the parameters to execute transaction
type ExecuteTxRequest struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte
}

// ExecuteTxOpts allows the user to specify more advanced options
type ExecuteTxOpts struct {
	Notifier           chan ExecuteTxResponse // async
	TxFilter           ExecuteTxFilter
	ProposalProcessors []ProposalProcessor // targets
	Timeout            time.Duration
}

// ExecuteTxFilter allows the user to inspect/modify response before commit
type ExecuteTxFilter interface {
	// process transaction proposal response (there will be no commit if an error is returned)
	ProcessTxProposalResponse(txProposalResponse []*TransactionProposalResponse) ([]*TransactionProposalResponse, error)
}

// Registration is a handle that is returned from a successful Register Chaincode Event.
// This handle should be used in Unregister in order to unregister the event.
type Registration interface {
}

// CCEvent contains the data for a chaincocde event
type CCEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
	Payload     []byte
}

// ChannelClient ...
/*
 * A channel client instance provides a handler to interact with peers on specified channel.
 * An application that requires interaction with multiple channels should create a separate
 * instance of the channel client for each channel. Channel client supports non-admin functions only.
 *
 * Each Client instance maintains {@link Channel} instance representing channel and the associated
 * private ledgers.
 *
 */
type ChannelClient interface {

	// Query chaincode
	Query(request QueryRequest) ([]byte, error)

	// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
	QueryWithOpts(request QueryRequest, opt QueryOpts) ([]byte, error)

	// ExecuteTx execute transaction
	ExecuteTx(request ExecuteTxRequest) (TransactionID, error)

	// ExecuteTxWithOpts allows the user to provide options for transaction execution (sync vs async, etc.)
	ExecuteTxWithOpts(request ExecuteTxRequest, opt ExecuteTxOpts) (TransactionID, error)

	// RegisterChaincodeEvent registers chain code event
	// @param {chan bool} channel which receives event details when the event is complete
	// @returns {object}  object handle that should be used to unregister
	RegisterChaincodeEvent(notify chan<- *CCEvent, chainCodeID string, eventID string) Registration

	// UnregisterChaincodeEvent unregisters chain code event
	UnregisterChaincodeEvent(registration Registration) error

	// Close releases channel client resources (disconnects event hub etc.)
	Close() error
}
