/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmtclient enables resource management client
package resmgmtclient

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ResourceMgmtClient enables managing resources in Fabric network.
type ResourceMgmtClient struct {
	client    fab.FabricClient
	config    config.Config
	filter    resmgmt.TargetFilter
	discovery fab.DiscoveryService
}

// MSPFilter is default filter
type MSPFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *MSPFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// NewResourceMgmtClient returns a ResourceMgmtClient instance
func NewResourceMgmtClient(client fab.FabricClient, discovery fab.DiscoveryService, filter resmgmt.TargetFilter, config config.Config) (*ResourceMgmtClient, error) {

	if client.UserContext() == nil {
		return nil, errors.New("must provide client identity")
	}

	rcFilter := filter
	if rcFilter == nil {
		// Default target filter is based on user msp
		rcFilter = &MSPFilter{mspID: client.UserContext().MspID()}
	}

	resourceClient := &ResourceMgmtClient{client: client, discovery: discovery, filter: rcFilter, config: config}
	return resourceClient, nil
}

// JoinChannel allows for default peers to join existing channel. Default peers are selected by applying default filter to all network peers.
func (rc *ResourceMgmtClient) JoinChannel(channelID string) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	targets, err := rc.getDefaultTargets()
	if err != nil {
		return errors.WithMessage(err, "failed to get default targets for JoinChannel")
	}

	return rc.JoinChannelWithOpts(channelID, resmgmt.JoinChannelOpts{Targets: targets})
}

//JoinChannelWithOpts allows for customizing set of peers about to join the channel (specific peers or custom 'filtered' peers)
func (rc *ResourceMgmtClient) JoinChannelWithOpts(channelID string, opts resmgmt.JoinChannelOpts) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	targets, err := rc.calculateTargets(opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for JoinChannel")
	}

	txnid, err := rc.client.NewTxnID()
	if err != nil {
		return errors.WithMessage(err, "NewTxnID failed")
	}

	channel, err := rc.getChannel(channelID)
	if err != nil {
		return errors.WithMessage(err, "get channel failed")
	}

	genesisBlock, err := channel.GenesisBlock(&fab.GenesisBlockRequest{TxnID: txnid})
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	txnid2, err := rc.client.NewTxnID()
	if err != nil {
		return errors.WithMessage(err, "NewTxnID failed")
	}

	joinChannelRequest := &fab.JoinChannelRequest{
		Targets:      targets,
		GenesisBlock: genesisBlock,
		TxnID:        txnid2,
	}

	err = channel.JoinChannel(joinChannelRequest)
	if err != nil {
		return errors.WithMessage(err, "join channel failed")
	}

	return nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []fab.Peer, filter resmgmt.TargetFilter) []fab.Peer {

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

// helper method for calculating default targets
func (rc *ResourceMgmtClient) getDefaultTargets() ([]fab.Peer, error) {

	// Default targets are discovery peers
	peers, err := rc.discovery.GetPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to discover peers")
	}

	// Apply default filter to discovery peers
	targets := filterTargets(peers, rc.filter)

	return targets, nil

}

// calculateTargets calculates targets based on targets and filter
func (rc *ResourceMgmtClient) calculateTargets(peers []fab.Peer, filter resmgmt.TargetFilter) ([]fab.Peer, error) {

	if peers != nil && filter != nil {
		return nil, errors.New("If targets are provided, filter cannot be provided")
	}

	targets := peers
	targetFilter := filter

	var err error
	if targets == nil {
		// Retrieve targets from discovery
		targets, err = rc.discovery.GetPeers()
		if err != nil {
			return nil, err
		}

		if filter == nil {
			targetFilter = rc.filter
		}
	}

	if targetFilter != nil {
		targets = filterTargets(targets, targetFilter)
	}

	return targets, nil
}

// getChannel is helper method for creating channel. If channel is not configured it will use random orderer from global orderer configuration
func (rc *ResourceMgmtClient) getChannel(channelID string) (fab.Channel, error) {

	channel := rc.client.Channel(channelID)
	if channel != nil {
		return channel, nil
	}

	// Creating channel requires orderer information
	var orderers []config.OrdererConfig
	chCfg, err := rc.config.ChannelConfig(channelID)
	if err != nil {
		return nil, err
	}

	if chCfg == nil {
		orderers, err = rc.config.OrderersConfig()
	} else {
		orderers, err = rc.config.ChannelOrderers(channelID)
	}

	// Check if retrieving orderer configuration went ok
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to retrieve orderer configuration")
	}

	if len(orderers) == 0 {
		return nil, errors.Errorf("Must configure at least one order for channel and/or one orderer in the network")
	}

	channel, err = rc.client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	for _, ordererCfg := range orderers {
		orderer, err := orderer.NewOrdererFromConfig(&ordererCfg, rc.config)
		if err != nil {
			return nil, errors.WithMessage(err, "NewOrdererFromConfig failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}
