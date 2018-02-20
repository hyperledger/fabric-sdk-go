/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmtclient enables resource management client
package resmgmtclient

import (
	"io/ioutil"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ResourceMgmtClient enables managing resources in Fabric network.
type ResourceMgmtClient struct {
	provider          fab.ProviderContext
	identity          fab.IdentityContext
	discoveryProvider fab.DiscoveryProvider // used to get per channel discovery service(s)
	channelProvider   fab.ChannelProvider
	fabricProvider    api.FabricProvider
	discovery         fab.DiscoveryService // global discovery service (detects all peers on the network)
	resource          fab.Resource
	filter            resmgmt.TargetFilter
}

// MSPFilter is default filter
type MSPFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *MSPFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// Context holds the providers and services needed to create a ChannelClient.
type Context struct {
	fab.ProviderContext
	fab.IdentityContext
	DiscoveryProvider fab.DiscoveryProvider
	ChannelProvider   fab.ChannelProvider
	FabricProvider    api.FabricProvider
	Resource          fab.Resource
}

type fabContext struct {
	fab.ProviderContext
	fab.IdentityContext
}

// New returns a ResourceMgmtClient instance
func New(ctx Context, filter resmgmt.TargetFilter) (*ResourceMgmtClient, error) {

	rcFilter := filter
	if rcFilter == nil {
		// Default target filter is based on user msp
		if ctx.MspID() == "" {
			return nil, errors.New("mspID not available in user context")
		}

		rcFilter = &MSPFilter{mspID: ctx.MspID()}
	}

	// setup global discovery service
	discovery, err := ctx.DiscoveryProvider.NewDiscoveryService("")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create global discovery service")
	}

	resourceClient := ResourceMgmtClient{
		provider:          ctx,
		identity:          ctx,
		discoveryProvider: ctx.DiscoveryProvider,
		channelProvider:   ctx.ChannelProvider,
		fabricProvider:    ctx.FabricProvider,
		resource:          ctx.Resource,
		discovery:         discovery,
		filter:            rcFilter,
	}
	return &resourceClient, nil
}

// JoinChannel allows for peers to join existing channel with optional custom options (specific peers, filtered peers)
func (rc *ResourceMgmtClient) JoinChannel(channelID string, options ...resmgmt.Option) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	opts, err := rc.prepareResmgmtOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for JoinChannel")
	}

	targets, err := rc.calculateTargets(rc.discovery, opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for JoinChannel")
	}

	if len(targets) == 0 {
		return errors.New("No targets available")
	}

	// TODO: should the code to get orderers from sdk config be part of channel service?
	oConfig, err := rc.provider.Config().ChannelOrderers(channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to load orderer config")
	}
	if len(oConfig) == 0 {
		return errors.Errorf("no orderers are configured for channel %s", channelID)
	}

	// TODO: handle more than the first orderer.
	orderer, err := rc.fabricProvider.CreateOrdererFromConfig(&oConfig[0])
	if err != nil {
		return errors.WithMessage(err, "failed to create orderers from config")
	}

	genesisBlock, err := rc.resource.GenesisBlockFromOrderer(channelID, orderer)
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	joinChannelRequest := fab.JoinChannelRequest{
		Targets:      peersToTxnProcessors(targets),
		GenesisBlock: genesisBlock,
	}

	err = rc.resource.JoinChannel(joinChannelRequest)
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
	chaincodeQueryResponse, err := rc.resource.QueryInstalledChaincodes(peer)
	if err != nil {
		return false, err
	}

	logger.Debugf("isChaincodeInstalled: %v", chaincodeQueryResponse)

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == req.Name && chaincode.Version == req.Version && chaincode.Path == req.Path {
			return true, nil
		}
	}

	return false, nil
}

// InstallCC installs chaincode with optional custom options (specific peers, filtered peers)
func (rc *ResourceMgmtClient) InstallCC(req resmgmt.InstallCCRequest, options ...resmgmt.Option) ([]resmgmt.InstallCCResponse, error) {

	// For each peer query if chaincode installed. If cc is installed treat as success with message 'already installed'.
	// If cc is not installed try to install, and if that fails add to the list with error and peer name.

	err := checkRequiredInstallCCParams(req)
	if err != nil {
		return nil, err
	}

	opts, err := rc.prepareResmgmtOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for InstallCC")
	}

	//Default targets when targets are not provided in options
	if len(opts.Targets) == 0 {
		opts.Targets, err = rc.getDefaultTargets(rc.discovery)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get default targets for InstallCC")
		}
	}

	targets, err := rc.calculateTargets(rc.discovery, opts.Targets, opts.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for install cc")
	}

	if len(targets) == 0 {
		return nil, errors.New("No targets available for install cc")
	}

	responses := make([]resmgmt.InstallCCResponse, 0)
	var errs multi.Errors

	// Targets will be adjusted if cc has already been installed
	newTargets := make([]fab.Peer, 0)
	for _, target := range targets {
		installed, err := rc.isChaincodeInstalled(req, target)
		if err != nil {
			// Add to errors with unable to verify error message
			errs = append(errs, errors.Errorf("unable to verify if cc is installed on %s", target.URL()))
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
	transactionProposalResponse, _, err := rc.resource.InstallChaincode(icr)
	for _, v := range transactionProposalResponse {
		logger.Debugf("Install chaincode '%s' endorser '%s' returned ProposalResponse status:%v", req.Name, v.Endorser, v.Status)

		response := resmgmt.InstallCCResponse{Target: v.Endorser, Status: v.Status}
		responses = append(responses, response)
	}

	if err != nil {
		return responses, errors.WithMessage(err, "InstallChaincode failed")
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
func (rc *ResourceMgmtClient) InstantiateCC(channelID string, req resmgmt.InstantiateCCRequest, options ...resmgmt.Option) error {

	return rc.sendCCProposal(channel.InstantiateChaincode, channelID, req, options...)

}

// UpgradeCC upgrades chaincode  with optional custom options (specific peers, filtered peers, timeout)
func (rc *ResourceMgmtClient) UpgradeCC(channelID string, req resmgmt.UpgradeCCRequest, options ...resmgmt.Option) error {

	return rc.sendCCProposal(channel.UpgradeChaincode, channelID, resmgmt.InstantiateCCRequest(req), options...)

}

// sendCCProposal sends proposal for type  Instantiate, Upgrade
func (rc *ResourceMgmtClient) sendCCProposal(ccProposalType channel.ChaincodeProposalType, channelID string, req resmgmt.InstantiateCCRequest, options ...resmgmt.Option) error {

	if err := checkRequiredCCProposalParams(channelID, req); err != nil {
		return err
	}

	opts, err := rc.prepareResmgmtOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for cc proposal")
	}

	// per channel discovery service
	discovery, err := rc.discoveryProvider.NewDiscoveryService(channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel discovery service")
	}

	//Default targets when targets are not provided in options
	if len(opts.Targets) == 0 {
		opts.Targets, err = rc.getDefaultTargets(discovery)
		if err != nil {
			return errors.WithMessage(err, "failed to get default targets for cc proposal")
		}
	}

	targets, err := rc.calculateTargets(discovery, opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for cc proposal")
	}

	if len(targets) == 0 {
		return errors.New("No targets available for cc proposal")
	}

	// Get transactor on the channel to create and send the deploy proposal
	channelService, err := rc.channelProvider.NewChannelService(rc.identity, channelID)
	if err != nil {
		return errors.WithMessage(err, "Unable to get channel service")
	}
	transactor, err := channelService.Transactor()
	if err != nil {
		return errors.WithMessage(err, "get channel transactor failed")
	}

	// create a transaction proposal for chaincode deployment
	deployProposal := channel.ChaincodeDeployRequest(req)
	deployCtx := fabContext{
		ProviderContext: rc.provider,
		IdentityContext: rc.identity,
	}

	txid, err := txn.NewID(&deployCtx)
	if err != nil {
		return errors.WithMessage(err, "create transaction ID failed")
	}
	tp, err := channel.CreateChaincodeDeployProposal(txid, ccProposalType, channelID, deployProposal)
	if err != nil {
		return errors.WithMessage(err, "creating chaincode deploy transaction proposal failed")
	}

	// Process and send transaction proposal
	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return errors.WithMessage(err, "sending deploy transaction proposal failed")
	}

	eventHub, err := channelService.EventHub()
	if err != nil {
		return errors.WithMessage(err, "Unable to get EventHub")
	}
	if eventHub.IsConnected() == false {
		err := eventHub.Connect()
		if err != nil {
			return err
		}
		defer eventHub.Disconnect()
	}

	// Register for commit event
	statusNotifier := txn.RegisterStatus(tp.TxnID, eventHub)

	transactionRequest := fab.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: txProposalResponse,
	}
	if _, err = createAndSendTransaction(transactor, transactionRequest); err != nil {
		return errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	timeout := rc.provider.Config().TimeoutOrDefault(config.Execute)
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}

	select {
	case result := <-statusNotifier:
		if result.Error == nil {
			return nil
		}
		return errors.WithMessage(result.Error, "instantiateOrUpgradeCC failed")
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

//prepareResmgmtOpts Reads resmgmt.Opts from resmgmt.Option array
func (rc *ResourceMgmtClient) prepareResmgmtOpts(options ...resmgmt.Option) (resmgmt.Opts, error) {
	resmgmtOpts := resmgmt.Opts{}
	for _, option := range options {
		err := option(&resmgmtOpts)
		if err != nil {
			return resmgmtOpts, errors.WithMessage(err, "Failed to read resource management opts")
		}
	}
	return resmgmtOpts, nil
}

func createAndSendTransaction(sender fab.Sender, request fab.TransactionRequest) (*fab.TransactionResponse, error) {

	tx, err := sender.CreateTransaction(request)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}
	if transactionResponse.Err != nil {
		logger.Debugf("orderer %s failed (%s)", transactionResponse.Orderer, transactionResponse.Err.Error())
		return nil, errors.Wrap(transactionResponse.Err, "orderer failed")
	}

	return transactionResponse, nil
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

// SaveChannel creates or updates channel
func (rc *ResourceMgmtClient) SaveChannel(req resmgmt.SaveChannelRequest, options ...resmgmt.Option) error {

	opts, err := rc.prepareSaveChannelOpts(options...)
	if err != nil {
		return err
	}

	if req.ChannelID == "" || req.ChannelConfig == "" {
		return errors.New("must provide channel ID and channel config")
	}

	logger.Debugf("***** Saving channel: %s *****\n", req.ChannelID)

	// Signing user has to belong to one of configured channel organisations
	// In case that order org is one of channel orgs we can use context user
	signer := rc.identity
	if req.SigningIdentity != nil {
		// Retrieve custom signing identity here
		signer = req.SigningIdentity
	}

	if signer == nil {
		return errors.New("must provide signing user")
	}

	configTx, err := ioutil.ReadFile(req.ChannelConfig)
	if err != nil {
		return errors.WithMessage(err, "reading channel config file failed")
	}

	chConfig, err := resource.ExtractChannelConfig(configTx)
	if err != nil {
		return errors.WithMessage(err, "extracting channel config failed")
	}

	sigCtx := Context{
		IdentityContext: signer,
		ProviderContext: rc.provider,
	}
	configSignature, err := resource.CreateConfigSignature(&sigCtx, chConfig)
	if err != nil {
		return errors.WithMessage(err, "signing configuration failed")
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	// Figure out orderer configuration
	var ordererCfg *config.OrdererConfig
	if opts.OrdererID != "" {
		ordererCfg, err = rc.provider.Config().OrdererConfig(opts.OrdererID)
	} else {
		// Default is random orderer from configuration
		ordererCfg, err = rc.provider.Config().RandomOrdererConfig()
	}

	// Check if retrieving orderer configuration went ok
	if err != nil || ordererCfg == nil {
		return errors.Errorf("failed to retrieve orderer config: %s", err)
	}

	orderer, err := orderer.New(rc.provider.Config(), orderer.FromOrdererConfig(ordererCfg))
	if err != nil {
		return errors.WithMessage(err, "failed to create new orderer from config")
	}

	request := fab.CreateChannelRequest{
		Name:       req.ChannelID,
		Orderer:    orderer,
		Config:     chConfig,
		Signatures: configSignatures,
	}

	_, err = rc.resource.CreateChannel(request)
	if err != nil {
		return errors.WithMessage(err, "create channel failed")
	}

	return nil
}

//prepareSaveChannelOpts Reads chmgmt.Opts from chmgmt.Option array
func (rc *ResourceMgmtClient) prepareSaveChannelOpts(options ...resmgmt.Option) (resmgmt.Opts, error) {
	saveChannelOpts := resmgmt.Opts{}
	for _, option := range options {
		err := option(&saveChannelOpts)
		if err != nil {
			return saveChannelOpts, errors.WithMessage(err, "Failed to read save channel opts")
		}
	}
	return saveChannelOpts, nil
}
