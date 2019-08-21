/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package invoke provides the handlers for performing chaincode invocations.
package invoke

import (
	reqContext "context"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// CCFilter returns true if the given chaincode should be included
// in the invocation chain when computing endorsers.
type CCFilter func(ccID string) bool

// Opts allows the user to specify more advanced options
type Opts struct {
	Targets       []fab.Peer // targets
	TargetFilter  fab.TargetFilter
	TargetSorter  fab.TargetSorter
	Retry         retry.Opts
	BeforeRetry   retry.BeforeRetryHandler
	Timeouts      map[fab.TimeoutType]time.Duration
	ParentContext reqContext.Context //parent grpc context
	CCFilter      CCFilter
}

// Request contains the parameters to execute transaction
type Request struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte

	// InvocationChain contains meta-data that's used by some Selection Service implementations
	// to choose endorsers that satisfy the endorsement policies of all chaincodes involved
	// in an invocation chain (i.e. for CC-to-CC invocations).
	// Each chaincode may also be associated with a set of private data collection names
	// which are used by some Selection Services (e.g. Fabric Selection) to exclude endorsers
	// that do NOT have read access to the collections.
	// The invoked chaincode (specified by ChaincodeID) may optionally be added to the invocation
	// chain along with any collections, otherwise it may be omitted.
	InvocationChain []*fab.ChaincodeCall
}

//Response contains response parameters for query and execute transaction
type Response struct {
	Proposal         *fab.TransactionProposal
	Responses        []*fab.TransactionProposalResponse
	TransactionID    fab.TransactionID
	TxValidationCode pb.TxValidationCode
	ChaincodeStatus  int32
	Payload          []byte
}

//Handler for chaining transaction executions
type Handler interface {
	Handle(context *RequestContext, clientContext *ClientContext)
}

//ClientContext contains context parameters for handler execution
type ClientContext struct {
	CryptoSuite  core.CryptoSuite
	Discovery    fab.DiscoveryService
	Selection    fab.SelectionService
	Membership   fab.ChannelMembership
	Transactor   fab.Transactor
	EventService fab.EventService
}

//RequestContext contains request, opts, response parameters for handler execution
type RequestContext struct {
	Request         Request
	Opts            Opts
	Response        Response
	Error           error
	RetryHandler    retry.Handler
	Ctx             reqContext.Context
	SelectionFilter selectopts.PeerFilter
	PeerSorter      selectopts.PeerSorter
}
