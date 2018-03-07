/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ledger enables ability to query ledger in a Fabric network.
package ledger

import (
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

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
	context   context.Client
	discovery fab.DiscoveryService
	ledger    *channel.Ledger
	filter    TargetFilter
	chName    string
}

// MSPFilter is default filter
type MSPFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *MSPFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// New returns a Client instance.
func New(clientProvider context.ClientProvider, channelID string, opts ...ClientOption) (*Client, error) {

	clientContext, err := clientProvider()
	if err != nil {
		return nil, err
	}

	l, err := channel.NewLedger(clientContext, channelID)
	if err != nil {
		return nil, err
	}

	discoveryService, err := clientContext.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return nil, err
	}

	ledgerClient := Client{
		context:   clientContext,
		discovery: discoveryService,
		ledger:    l,
		chName:    channelID,
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
		if clientContext.MspID() == "" {
			return nil, errors.New("mspID not available in user context")
		}
		filter := &MSPFilter{mspID: clientContext.MspID()}
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

	responses, err := c.ledger.QueryInfo(peersToTxnProcessors(targets))
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

	responses, err := c.ledger.QueryBlockByHash(blockHash, peersToTxnProcessors(targets))
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
func (c *Client) QueryBlock(blockNumber int, options ...RequestOption) (*common.Block, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlock")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlock")
	}

	responses, err := c.ledger.QueryBlock(blockNumber, peersToTxnProcessors(targets))
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

	responses, err := c.ledger.QueryTransaction(transactionID, peersToTxnProcessors(targets))
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

	channelConfig, err := chconfig.New(c.context, c.chName, chconfig.WithPeers(targets), chconfig.WithMinResponses(opts.MinTargets))
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfig failed")
	}

	return channelConfig.Query()

}

//prepareRequestOpts Reads Opts from Option array
func (c *Client) prepareRequestOpts(options ...RequestOption) (Opts, error) {
	opts := Opts{}
	for _, option := range options {
		err := option(&opts)
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
func (c *Client) calculateTargets(opts Opts) ([]fab.Peer, error) {

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
