/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmtclient enables resource management client
package resmgmtclient

import (
	"time"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ResourceMgmtClient enables managing resources in Fabric network.
type ResourceMgmtClient struct {
	client    fab.FabricClient
	config    config.Config
	filter    resmgmt.TargetFilter
	discovery fab.DiscoveryService  // global discovery service (detects all peers on the network)
	provider  fab.DiscoveryProvider // used to get per channel discovery service(s)
}

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

// MSPFilter is default filter
type MSPFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *MSPFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// NewResourceMgmtClient returns a ResourceMgmtClient instance
func NewResourceMgmtClient(client fab.FabricClient, provider fab.DiscoveryProvider, filter resmgmt.TargetFilter, config config.Config) (*ResourceMgmtClient, error) {

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

	// setup global discovery service
	discovery, err := provider.NewDiscoveryService("")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create global discovery service")
	}

	resourceClient := &ResourceMgmtClient{client: client, discovery: discovery, provider: provider, filter: rcFilter, config: config}
	return resourceClient, nil
}

// JoinChannel allows for default peers to join existing channel. Default peers are selected by applying default filter to all network peers.
func (rc *ResourceMgmtClient) JoinChannel(channelID string) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	targets, err := rc.getDefaultTargets(rc.discovery)
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

	targets, err := rc.calculateTargets(rc.discovery, opts.Targets, opts.TargetFilter)
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
func (rc *ResourceMgmtClient) getDefaultTargets(discovery fab.DiscoveryService) ([]fab.Peer, error) {

	// Default targets are discovery peers
	peers, err := discovery.GetPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to discover peers")
	}

	// Apply default filter to discovery peers
	targets := filterTargets(peers, rc.filter)

	return targets, nil

}

// calculateTargets calculates targets based on targets and filter
func (rc *ResourceMgmtClient) calculateTargets(discovery fab.DiscoveryService, peers []fab.Peer, filter resmgmt.TargetFilter) ([]fab.Peer, error) {

	if peers != nil && filter != nil {
		return nil, errors.New("If targets are provided, filter cannot be provided")
	}

	targets := peers
	targetFilter := filter

	var err error
	if targets == nil {
		// Retrieve targets from discovery
		targets, err = discovery.GetPeers()
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

// isChaincodeInstalled verify if chaincode is installed on peer
func (rc *ResourceMgmtClient) isChaincodeInstalled(req resmgmt.InstallCCRequest, peer fab.Peer) (bool, error) {
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

	targets, err := rc.getDefaultTargets(rc.discovery)
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

	targets, err := rc.calculateTargets(rc.discovery, opts.Targets, opts.TargetFilter)
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
		installed, err := rc.isChaincodeInstalled(req, target)
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
		return errors.New("Chaincode name, version, path and chaincode package are required")
	}
	return nil
}

// InstantiateCC instantiates chaincode using default settings
func (rc *ResourceMgmtClient) InstantiateCC(channelID string, req resmgmt.InstantiateCCRequest) error {

	if err := checkRequiredCCProposalParams(channelID, req); err != nil {
		return err
	}

	// per channel discovery service
	discovery, err := rc.provider.NewDiscoveryService(channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel discovery service")
	}

	targets, err := rc.getDefaultTargets(discovery)
	if err != nil {
		return errors.WithMessage(err, "failed to get default targets for InstantiateCC")
	}

	if len(targets) == 0 {
		return errors.New("No default targets available for instantiate cc")
	}

	return rc.InstantiateCCWithOpts(channelID, req, resmgmt.InstantiateCCOpts{Targets: targets})
}

// InstantiateCCWithOpts instantiates chaincode with custom options
func (rc *ResourceMgmtClient) InstantiateCCWithOpts(channelID string, req resmgmt.InstantiateCCRequest, opts resmgmt.InstantiateCCOpts) error {

	return rc.sendCCProposalWithOpts(Instantiate, channelID, req, opts)

}

// UpgradeCC upgrades chaincode using default settings
func (rc *ResourceMgmtClient) UpgradeCC(channelID string, req resmgmt.UpgradeCCRequest) error {

	if err := checkRequiredCCProposalParams(channelID, resmgmt.InstantiateCCRequest(req)); err != nil {
		return err
	}

	// per channel discovery service
	discovery, err := rc.provider.NewDiscoveryService(channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel discovery service")
	}

	targets, err := rc.getDefaultTargets(discovery)
	if err != nil {
		return errors.WithMessage(err, "failed to get default targets for UpgradeCC")
	}

	if len(targets) == 0 {
		return errors.New("No default targets available for upgrade cc")
	}

	return rc.UpgradeCCWithOpts(channelID, req, resmgmt.UpgradeCCOpts{Targets: targets})
}

// UpgradeCCWithOpts upgrades chaincode with custom options
func (rc *ResourceMgmtClient) UpgradeCCWithOpts(channelID string, req resmgmt.UpgradeCCRequest, opts resmgmt.UpgradeCCOpts) error {

	return rc.sendCCProposalWithOpts(Upgrade, channelID, resmgmt.InstantiateCCRequest(req), resmgmt.InstantiateCCOpts(opts))

}

// InstantiateCCWithOpts instantiates chaincode with custom options
func (rc *ResourceMgmtClient) sendCCProposalWithOpts(ccProposalType CCProposalType, channelID string, req resmgmt.InstantiateCCRequest, opts resmgmt.InstantiateCCOpts) error {

	if err := checkRequiredCCProposalParams(channelID, req); err != nil {
		return err
	}

	// per channel discovery service
	discovery, err := rc.provider.NewDiscoveryService(channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel discovery service")
	}

	targets, err := rc.calculateTargets(discovery, opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for cc proposal")
	}

	if len(targets) == 0 {
		return errors.New("No targets available for cc proposal")
	}

	channel, err := rc.getChannel(channelID)
	if err != nil {
		return errors.WithMessage(err, "get channel failed")
	}

	var txProposalResponse []*apitxn.TransactionProposalResponse
	var txID apitxn.TransactionID

	switch ccProposalType {

	case Instantiate:
		txProposalResponse, txID, err = channel.SendInstantiateProposal(req.Name,
			req.Args, req.Path, req.Version, req.Policy, peer.PeersToTxnProcessors(targets))
		if err != nil {
			return errors.Wrap(err, "send instantiate chaincode proposal failed")
		}
	case Upgrade:
		txProposalResponse, txID, err = channel.SendUpgradeProposal(req.Name,
			req.Args, req.Path, req.Version, req.Policy, peer.PeersToTxnProcessors(targets))
		if err != nil {
			return errors.Wrap(err, "send upgrade chaincode proposal failed")
		}
	default:
		return errors.Errorf("chaincode proposal type %d not supported", ccProposalType)
	}

	for _, v := range txProposalResponse {
		if v.Err != nil {
			return errors.WithMessage(v.Err, "cc proposal failed")
		}
	}

	eventHub, err := rc.getEventHub(channelID)
	if err != nil {
		return errors.WithMessage(err, "get event hub failed")
	}

	if eventHub.IsConnected() == false {
		err := eventHub.Connect()
		if err != nil {
			return err
		}
		defer eventHub.Disconnect()
	}

	// Register for commit event
	chcode := internal.RegisterTxEvent(txID, eventHub)

	if _, err = internal.CreateAndSendTransaction(channel, txProposalResponse); err != nil {
		return errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	timeout := rc.config.TimeoutOrDefault(config.ExecuteTx)
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}

	select {
	case code := <-chcode:
		if code == pb.TxValidationCode_VALID {
			return nil
		}
		return errors.Errorf("instantiateOrUpgradeCC received tx validation code %s", code)
	case <-time.After(timeout):
		return errors.New("instantiateOrUpgradeCC timeout")
	}

}

func checkRequiredCCProposalParams(channelID string, req resmgmt.InstantiateCCRequest) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	if req.Name == "" || req.Version == "" || req.Path == "" || req.Policy == nil {
		return errors.New("Chaincode name, version, path and policy are required")
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

func (rc *ResourceMgmtClient) getEventHub(channelID string) (*events.EventHub, error) {

	peerConfig, err := rc.config.ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "read configuration for channel peers failed")
	}

	var eventSource *config.ChannelPeer
	for _, p := range peerConfig {
		if p.EventSource && p.MspID == rc.client.UserContext().MspID() {
			eventSource = &p
			break
		}
	}

	if eventSource == nil {
		return nil, errors.New("unable to find event source for channel")
	}

	// Event source found, create event hub
	return events.NewEventHubFromConfig(rc.client, &eventSource.PeerConfig)

}
