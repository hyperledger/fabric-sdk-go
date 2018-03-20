/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package invoke provides the handlers for performing chaincode invocations.
package invoke

import (
	reqContext "context"
	"time"

	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Opts allows the user to specify more advanced options
type Opts struct {
	Targets       []fab.Peer // targets
	TargetFilter  fab.TargetFilter
	Retry         retry.Opts
	Timeouts      map[core.TimeoutType]time.Duration
	ParentContext reqContext.Context //parent grpc context
}

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
}
