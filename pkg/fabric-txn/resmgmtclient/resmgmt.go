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
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
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
		if client.UserContext().MspID() == "" {
			return nil, errors.New("mspID not available in user context")
		}

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

	if len(targets) == 0 {
		return errors.New("No default targets available")
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

	if len(targets) == 0 {
		return errors.New("No targets available")
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

// IsChaincodeInstalled verify if chaincode is installed on peer
func (rc *ResourceMgmtClient) IsChaincodeInstalled(req resmgmt.InstallCCRequest, peer fab.Peer) (bool, error) {
	chaincodeQueryResponse, err := rc.client.QueryInstalledChaincodes(peer)
	if err != nil {
		return false, err
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == req.Name && chaincode.Version == req.Version && chaincode.Path == req.Path {
			return true, nil
		}
	}

	return false, nil
}

// InstallCC - install chaincode
func (rc *ResourceMgmtClient) InstallCC(req resmgmt.InstallCCRequest) ([]resmgmt.InstallCCResponse, error) {

	if err := checkRequiredInstallCCParams(req); err != nil {
		return nil, err
	}

	targets, err := rc.getDefaultTargets()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get default targets for InstallCC")
	}

	if len(targets) == 0 {
		return nil, errors.New("No default targets available for install cc")
	}

	return rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: targets})
}

// InstallCCWithOpts installs chaincode with custom options
func (rc *ResourceMgmtClient) InstallCCWithOpts(req resmgmt.InstallCCRequest, opts resmgmt.InstallCCOpts) ([]resmgmt.InstallCCResponse, error) {

	// For each peer query if chaincode installed. If cc is installed treat as success with message 'already installed'.
	// If cc is not installed try to install, and if that fails add to the list with error and peer name.

	err := checkRequiredInstallCCParams(req)
	if err != nil {
		return nil, err
	}

	targets, err := rc.calculateTargets(opts.Targets, opts.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for install cc")
	}

	if len(targets) == 0 {
		return nil, errors.New("No targets available for install cc")
	}

	responses := make([]resmgmt.InstallCCResponse, 0)

	// Targets will be adjusted if cc has already been installed
	newTargets := make([]fab.Peer, 0)

	for _, target := range targets {
		installed, err := rc.IsChaincodeInstalled(req, target)
		if err != nil {
			// Add to responses with unable to verify error message
			response := resmgmt.InstallCCResponse{Target: target.URL(), Err: errors.Errorf("unable to verify if cc is installed on %s", target.URL())}
			responses = append(responses, response)
			continue
		}
		if installed {
			// Nothing to do - add info message to response
			response := resmgmt.InstallCCResponse{Target: target.URL(), Info: "already installed"}
			responses = append(responses, response)
		} else {
			// Not installed - add for processing
			newTargets = append(newTargets, target)
		}
	}

	if len(newTargets) == 0 {
		// CC is already installed on all targets and/or
		// we are unable to verify if cc is installed on target(s)
		return responses, nil
	}

	icr := fab.InstallChaincodeRequest{Name: req.Name, Path: req.Path, Version: req.Version, Package: req.Package, Targets: peer.PeersToTxnProcessors(newTargets)}
	transactionProposalResponse, _, err := rc.client.InstallChaincode(icr)
	if err != nil {
		return nil, errors.WithMessage(err, "InstallChaincode failed")
	}

	for _, v := range transactionProposalResponse {

		logger.Infof("Install chaincode '%s' endorser '%s' returned ProposalResponse status:%v, error:'%s'", req.Name, v.Endorser, v.Status, v.Err)

		response := resmgmt.InstallCCResponse{Target: v.Endorser, Status: v.Status, Err: v.Err}
		responses = append(responses, response)
	}

	return responses, nil

}

func checkRequiredInstallCCParams(req resmgmt.InstallCCRequest) error {
	if req.Name == "" || req.Version == "" || req.Path == "" || req.Package == nil {
		return errors.New("Chaincode name(ID), version, path and chaincode package are required")
	}
	return nil
}

// getChannel is helper method for instantiating channel. If channel is not configured it will use random orderer from global orderer configuration
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
