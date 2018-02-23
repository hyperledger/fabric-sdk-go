/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// CCEvent contains the data for a chaincocde event
type CCEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
	Payload     []byte
}

// Registration is a handle that is returned from a successful Register Chaincode Event.
// This handle should be used in Unregister in order to unregister the event.
type Registration interface {
}

// Opts allows the user to specify more advanced options
type Opts struct {
	ProposalProcessors []fab.ProposalProcessor // targets
	Timeout            time.Duration
	Retry              retry.Opts
}

//Option func for each Opts argument
type Option func(opts *Opts) error

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
	TransactionID    fab.TransactionID
	TxValidationCode pb.TxValidationCode
	Proposal         *fab.TransactionProposal
	Responses        []*fab.TransactionProposalResponse
}

//Handler for chaining transaction executions
type Handler interface {
	Handle(context *RequestContext, clientContext *ClientContext)
}

//ClientContext contains context parameters for handler execution
type ClientContext struct {
	CryptoSuite core.CryptoSuite
	Discovery   fab.DiscoveryService
	Selection   fab.SelectionService
	Channel     fab.Channel // TODO: this should be removed when we have MSP split out.
	Transactor  fab.Transactor
	EventHub    fab.EventHub
}

//RequestContext contains request, opts, response parameters for handler execution
type RequestContext struct {
	Request      Request
	Opts         Opts
	Response     Response
	Error        error
	RetryHandler retry.Handler
}

//WithTimeout encapsulates time.Duration to Option
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Opts) error {
		opts.Timeout = timeout
		return nil
	}
}

//WithProposalProcessor encapsulates ProposalProcessors to Option
func WithProposalProcessor(proposalProcessors ...fab.ProposalProcessor) Option {
	return func(opts *Opts) error {
		opts.ProposalProcessors = proposalProcessors
		return nil
	}
}

// WithRetry option to configure retries
func WithRetry(opt retry.Opts) Option {
	return func(opts *Opts) error {
		opts.Retry = opt
		return nil
	}
}
