/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"time"

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

// opts allows the user to specify more advanced options
type opts struct {
	ProposalProcessors []fab.ProposalProcessor // targets
	Timeout            time.Duration
	Retry              retry.Opts
}

//Option func for each Opts argument
type Option func(opts *opts) error

// Request contains the parameters to query and execute an invocation transaction
type Request struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte
}

//Response contains response parameters for query and execute an invocation transaction
type Response struct {
	Payload          []byte
	TransactionID    fab.TransactionID
	TxValidationCode pb.TxValidationCode
	Proposal         *fab.TransactionProposal
	Responses        []*fab.TransactionProposalResponse
}

//WithTimeout encapsulates time.Duration to Option
func WithTimeout(timeout time.Duration) Option {
	return func(o *opts) error {
		o.Timeout = timeout
		return nil
	}
}

//WithProposalProcessor encapsulates ProposalProcessors to Option
func WithProposalProcessor(proposalProcessors ...fab.ProposalProcessor) Option {
	return func(o *opts) error {
		o.ProposalProcessors = proposalProcessors
		return nil
	}
}

// WithRetry option to configure retries
func WithRetry(retryOpt retry.Opts) Option {
	return func(o *opts) error {
		o.Retry = retryOpt
		return nil
	}
}
