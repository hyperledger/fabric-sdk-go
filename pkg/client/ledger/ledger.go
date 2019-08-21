/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ledger enables ledger queries on specified channel on a Fabric network.
// An application that requires ledger queries from multiple channels should create a separate
// instance of the ledger client for each channel. Ledger client supports the following queries:
// QueryInfo, QueryBlock, QueryBlockByHash,  QueryBlockByTxID, QueryTransaction and QueryConfig.
//
//  Basic Flow:
//  1) Prepare channel context
//  2) Create ledger client
//  3) Query ledger
package ledger

import (
	reqContext "context"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/filter"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"

	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/pkg/errors"
)

// Client enables ledger queries on a Fabric network.
type Client struct {
	ctx       context.Channel
	filter    fab.TargetFilter
	ledger    *channel.Ledger
	verifier  channel.ResponseVerifier
	discovery fab.DiscoveryService
}

// mspFilter is default filter
type mspFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// New returns a ledger client instance. A ledger client instance provides a handler to query various info on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the ledger client for each channel. Ledger client supports specific queries only.
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, err
	}

	if channelContext.ChannelService() == nil {
		return nil, errors.New("channel service not initialized")
	}

	membership, err := channelContext.ChannelService().Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	ledger, err := channel.NewLedger(channelContext.ChannelID())
	if err != nil {
		return nil, err
	}

	ledgerFilter := filter.NewEndpointFilter(channelContext, filter.LedgerQuery)

	discoveryService, err := channelContext.ChannelService().Discovery()
	if err != nil {
		return nil, err
	}

	// Apply filter to discovery service
	discovery := discovery.NewDiscoveryFilterService(discoveryService, ledgerFilter)

	ledgerClient := Client{
		ctx:       channelContext,
		ledger:    ledger,
		verifier:  &verifier.Signature{Membership: membership},
		discovery: discovery,
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
		if channelContext.Identifier().MSPID == "" {
			return nil, errors.New("mspID not available in user context")
		}
		filter := &mspFilter{mspID: channelContext.Identifier().MSPID}
		ledgerClient.filter = filter
	}

	return &ledgerClient, nil
}

// QueryInfo queries for various useful blockchain information on this channel such as block height and current block hash.
//  Parameters:
//  options are optional request options
//
//  Returns:
//  blockchain information
func (c *Client) QueryInfo(options ...RequestOption) (*fab.BlockchainInfoResponse, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryInfo failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	responses, err := c.ledger.QueryInfo(reqCtx, peersToTxnProcessors(targets), c.verifier)
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "QueryInfo failed")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("Number of responses %d is less than MinTargets %d. Targets: %v, Error: %s", len(responses), opts.MinTargets, targets, err)
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

	return response, nil
}

// QueryBlockByHash queries the ledger for block by block hash.
//  Parameters:
//  blockHash is required block hash
//  options hold optional request options
//
//  Returns:
//  block information
func (c *Client) QueryBlockByHash(blockHash []byte, options ...RequestOption) (*common.Block, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlockByHash failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	responses, err := c.ledger.QueryBlockByHash(reqCtx, blockHash, peersToTxnProcessors(targets), c.verifier)
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "QueryBlockByHash failed")
	}

	return matchBlockData(responses, opts.MinTargets)
}

// QueryBlockByTxID queries for block which contains a transaction.
//  Parameters:
//  txID is required transaction ID
//  options hold optional request options
//
//  Returns:
//  block information
func (c *Client) QueryBlockByTxID(txID fab.TransactionID, options ...RequestOption) (*common.Block, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlockByTxID failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	responses, err := c.ledger.QueryBlockByTxID(reqCtx, txID, peersToTxnProcessors(targets), c.verifier)
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "QueryBlockByTxID failed")
	}

	return matchBlockData(responses, opts.MinTargets)
}

// QueryBlock queries the ledger for Block by block number.
//  Parameters:
//  blockNumber is required block number(ID)
//  options hold optional request options
//
//  Returns:
//  block information
func (c *Client) QueryBlock(blockNumber uint64, options ...RequestOption) (*common.Block, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlock failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	responses, err := c.ledger.QueryBlock(reqCtx, blockNumber, peersToTxnProcessors(targets), c.verifier)
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "QueryBlock failed")
	}

	return matchBlockData(responses, opts.MinTargets)
}

func (c *Client) prepareRequestParams(options ...RequestOption) ([]fab.Peer, *requestOptions, error) {
	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get opts")
	}

	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to determine target peers")
	}

	return targets, &opts, nil
}

func matchBlockData(responses []*common.Block, minTargets int) (*common.Block, error) {
	if len(responses) < minTargets {
		return nil, errors.Errorf("Number of responses %d is less than MinTargets %d", len(responses), minTargets)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// Block data has to match
		if !proto.Equal(response.Data, r.Data) {
			return nil, errors.New("Block data does not match")
		}
	}

	return response, nil

}

// QueryTransaction queries the ledger for processed transaction by transaction ID.
//  Parameters:
//  txID is required transaction ID
//  options hold optional request options
//
//  Returns:
//  processed transaction information
func (c *Client) QueryTransaction(transactionID fab.TransactionID, options ...RequestOption) (*pb.ProcessedTransaction, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryTransaction failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	responses, err := c.ledger.QueryTransaction(reqCtx, transactionID, peersToTxnProcessors(targets), c.verifier)
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "QueryTransaction failed")
	}

	if len(responses) < opts.MinTargets {
		return nil, errors.Errorf("QueryTransaction: Number of responses %d is less than MinTargets %d", len(responses), opts.MinTargets)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// All payloads have to match
		if !proto.Equal(response, r) {
			return nil, errors.New("Payloads for QueryTransaction do not match")
		}
	}

	return response, nil
}

// QueryConfig queries for channel configuration.
//  Parameters:
//  options hold optional request options
//
//  Returns:
//  channel configuration information
func (c *Client) QueryConfig(options ...RequestOption) (fab.ChannelCfg, error) {

	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfig failed to prepare request parameters")
	}
	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	channelConfig, err := chconfig.New(c.ctx.ChannelID(), chconfig.WithPeers(targets), chconfig.WithMinResponses(opts.MinTargets))
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfig failed")
	}

	return channelConfig.Query(reqCtx)
}

// QueryConfigBlock returns the current configuration block for the specified channel.
func (c *Client) QueryConfigBlock(options ...RequestOption) (*common.Block, error) {
	targets, opts, err := c.prepareRequestParams(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfigBlock failed to prepare request parameters")
	}

	reqCtx, cancel := c.createRequestContext(opts)
	defer cancel()

	return c.ledger.QueryConfigBlock(reqCtx, peersToTxnProcessors(targets), c.verifier)
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
		targets, err = c.discovery.GetPeers()
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
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
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
		opts.Timeouts = make(map[fab.TimeoutType]time.Duration)
	}

	if opts.Timeouts[fab.PeerResponse] == 0 {
		opts.Timeouts[fab.PeerResponse] = c.ctx.EndpointConfig().Timeout(fab.PeerResponse)
	}

	return contextImpl.NewRequest(c.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.PeerResponse]), contextImpl.WithParent(opts.ParentContext))
}

// filterTargets is helper method to filter peers
func filterTargets(peers []fab.Peer, filter fab.TargetFilter) []fab.Peer {

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
