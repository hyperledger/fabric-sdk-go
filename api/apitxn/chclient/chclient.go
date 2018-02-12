/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Request contains the parameters to execute transaction
type Request struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte
}

//Response contains response parameters for query and execute transaction
type Response struct {
	Payload          []byte
	TransactionID    apifabclient.TransactionID
	TxValidationCode pb.TxValidationCode
	Responses        []*apifabclient.TransactionProposalResponse
}

// Opts allows the user to specify more advanced options
type Opts struct {
	ProposalProcessors []apifabclient.ProposalProcessor // targets
	Timeout            time.Duration
	Retry              retry.Opts
}

//Option func for each Opts argument
type Option func(opts *Opts) error

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

	// Query chaincode with request and optional options provided
	Query(request Request, opts ...Option) (Response, error)

	// Execute execute transaction with request and optional options provided
	Execute(request Request, opts ...Option) (Response, error)

	// InvokeHandler invokes the given handler with the given request and optional options provided
	InvokeHandler(handler Handler, request Request, options ...Option) (Response, error)

	// RegisterChaincodeEvent registers chain code event
	// @param {chan bool} channel which receives event details when the event is complete
	// @returns {object}  object handle that should be used to unregister
	RegisterChaincodeEvent(notify chan<- *CCEvent, chainCodeID string, eventID string) (Registration, error)

	// UnregisterChaincodeEvent unregisters chain code event
	UnregisterChaincodeEvent(registration Registration) error

	// Close releases channel client resources (disconnects event hub etc.)
	Close() error
}
