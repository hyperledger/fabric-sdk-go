/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ledger enables ability to query ledger in a Fabric network.
package ledger

import (
	reqContext "context"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

const (
	defaultHandlerTimeout = time.Second * 10
)

// Client enables ledger queries on a Fabric network.
//
// A ledger client instance provides a handler to query various info on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the ledger client for each channel. Ledger client supports specific queries only.
type Client struct {
	ctx    context.Channel
	filter TargetFilter
	ledger *channel.Ledger
}

// mspFilter is default filter
type mspFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// New returns a Client instance.
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, err
	}

	ledger, err := channel.NewLedger(channelContext.ChannelID())
	if err != nil {
		return nil, err
	}

	ledgerClient := Client{
		ctx:    channelContext,
		ledger: ledger,
	}

	for _, opt := range opts {
		err := opt(&ledgerClient)
		if err != nil {
			return nil, err
		}
	}

	// check if target filter was set - if not set the default
	if ledgerClient.filter == nil {
		// Default target filter is based on user msp
		if channelContext.MSPID() == "" {
			return nil, errors.New("mspID not available in user context")
		}
		filter := &mspFilter{mspID: channelContext.MSPID()}
		ledgerClient.filter = filter
	}

	return &ledgerClient, nil
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
func (c *Client) QueryInfo(options ...RequestOption) (*fab.BlockchainInfoResponse, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlockByHash")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlockByHash")
	}

	reqCtx, cancel := c.createRequestContext(&opts)
	defer cancel()

	responses, err := c.ledger.QueryInfo(reqCtx, peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryBlockByHash")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("Number of responses %d is less than MinTargets %d. Targets: %v, Error: %v", len(responses), opts.MinTargets, targets, err)
	}

	response := responses[0]
	maxHeight := response.BCI.Height
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// Match one with highest block height,
		if r.BCI.Height > maxHeight {
			response = r
			maxHeight = r.BCI.Height
		}

	}

	return response, err
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to specified targets.
// Returns the block.
func (c *Client) QueryBlockByHash(blockHash []byte, options ...RequestOption) (*common.Block, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlockByHash")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlockByHash")
	}

	reqCtx, cancel := c.createRequestContext(&opts)
	defer cancel()

	responses, err := c.ledger.QueryBlockByHash(reqCtx, blockHash, peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryBlockByHash")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("QueryBlockByHash: Number of responses %d is less than MinTargets %d", len(responses), opts.MinTargets)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// All payloads have to match
		if !proto.Equal(response.Data, r.Data) {
			return nil, errors.New("Payloads for QueryBlockByHash do not match")
		}
	}

	return response, err
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to specified targets.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func (c *Client) QueryBlock(blockNumber uint64, options ...RequestOption) (*common.Block, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlock")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlock")
	}

	reqCtx, cancel := c.createRequestContext(&opts)
	defer cancel()

	responses, err := c.ledger.QueryBlock(reqCtx, blockNumber, peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryBlock")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("QueryBlock: Number of responses %d is less than MinTargets %d", len(responses), opts.MinTargets)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// TODO: Signature validation

		// All payloads have to match
		if !proto.Equal(response.Data, r.Data) {
			return nil, errors.New("Payloads for QueryBlock do not match")
		}
	}

	return response, err
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to specified targets.
// Returns the ProcessedTransaction information containing the transaction.
func (c *Client) QueryTransaction(transactionID fab.TransactionID, options ...RequestOption) (*pb.ProcessedTransaction, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryTransaction")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryTransaction")
	}

	reqCtx, cancel := c.createRequestContext(&opts)
	defer cancel()

	responses, err := c.ledger.QueryTransaction(reqCtx, transactionID, peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryTransaction")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("QueryTransaction: Number of responses %d is less than MinTargets %d", len(responses), opts.MinTargets)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// TODO: Signature validation

		// All payloads have to match
		if !proto.Equal(response, r) {
			return nil, errors.New("Payloads for QueryBlockByHash do not match")
		}
	}

	return response, err
}

// QueryConfig config returns channel configuration
func (c *Client) QueryConfig(options ...RequestOption) (fab.ChannelCfg, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryConfig")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryConfig")
	}

	channelConfig, err := chconfig.New(c.ctx.ChannelID(), chconfig.WithPeers(targets), chconfig.WithMinResponses(opts.MinTargets))
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfig failed")
	}

	reqCtx, cancel := c.createRequestContext(&opts)
	defer cancel()

	return channelConfig.Query(reqCtx)

}

//prepareRequestOpts Reads Opts from Option array
func (c *Client) prepareRequestOpts(options ...RequestOption) (requestOptions, error) {
	opts := requestOptions{}
	for _, option := range options {
		err := option(c.ctx, &opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read request opts")
		}
	}

	// Set defaults for max targets
	if opts.MaxTargets == 0 {
		opts.MaxTargets = maxTargets
	}

	// Set defaults for min targets/matches
	if opts.MinTargets == 0 {
		opts.MinTargets = minTargets
	}

	if opts.MinTargets > opts.MaxTargets {
		opts.MaxTargets = opts.MinTargets
	}

	return opts, nil
}

// calculateTargets calculates targets based on targets and filter
func (c *Client) calculateTargets(opts requestOptions) ([]fab.Peer, error) {

	if opts.Targets != nil && opts.TargetFilter != nil {
		return nil, errors.New("If targets are provided, filter cannot be provided")
	}

	targets := opts.Targets
	targetFilter := opts.TargetFilter

	var err error
	if targets == nil {
		// Retrieve targets from discovery
		targets, err = c.ctx.DiscoveryService().GetPeers()
		if err != nil {
			return nil, err
		}

		if targetFilter == nil {
			targetFilter = c.filter
		}
	}

	if targetFilter != nil {
		targets = filterTargets(targets, targetFilter)
	}

	if len(targets) == 0 {
		return nil, errors.New("No targets available")
	}

	if len(targets) < opts.MinTargets {
		return nil, errors.Errorf("Error getting minimum number of targets. %d available, %d required", len(targets), opts.MinTargets)
	}

	// Calculate number of targets required
	numOfTargets := opts.MaxTargets
	if len(targets) < opts.MaxTargets {
		numOfTargets = len(targets)
	}

	// Shuffle to randomize
	shuffle(targets)

	return targets[:numOfTargets], nil
}

//createRequestContext creates request context for grpc
func (c *Client) createRequestContext(opts *requestOptions) (reqContext.Context, reqContext.CancelFunc) {

	if opts.Timeouts == nil {
		opts.Timeouts = make(map[core.TimeoutType]time.Duration)
	}

	if opts.Timeouts[core.PeerResponse] == 0 {
		opts.Timeouts[core.PeerResponse] = c.ctx.Config().TimeoutOrDefault(core.PeerResponse)
	}

	return contextImpl.NewRequest(c.ctx, contextImpl.WithTimeout(opts.Timeouts[core.PeerResponse]), contextImpl.WithParent(opts.ParentContext))
}

// filterTargets is helper method to filter peers
func filterTargets(peers []fab.Peer, filter TargetFilter) []fab.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []fab.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

func shuffle(a []fab.Peer) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
