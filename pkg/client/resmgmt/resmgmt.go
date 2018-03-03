/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt enables ability to update resources in a Fabric network.
package resmgmt

import (
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer fab.Peer) bool
}

// InstallCCRequest contains install chaincode request parameters
type InstallCCRequest struct {
	Name    string
	Path    string
	Version string
	Package *api.CCPackage
}

// InstallCCResponse contains install chaincode response status
type InstallCCResponse struct {
	Target string
	Status int32
	Info   string
}

// InstantiateCCRequest contains instantiate chaincode request parameters
type InstantiateCCRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

// UpgradeCCRequest contains upgrade chaincode request parameters
type UpgradeCCRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

//Opts contains options for operations performed by ResourceMgmtClient
type Opts struct {
	Targets      []fab.Peer    // target peers
	TargetFilter TargetFilter  // target filter
	Timeout      time.Duration //timeout options for instantiate and upgrade CC
	OrdererID    string        // use specific orderer
}

//SaveChannelRequest used to save channel request
type SaveChannelRequest struct {
	// Channel Name (ID)
	ChannelID string
	// Path to channel configuration file
	ChannelConfig string
	// User that signs channel configuration
	SigningIdentity context.Identity
}

//RequestOption func for each Opts argument
type RequestOption func(opts *Opts) error

var logger = logging.NewLogger("fabric_sdk_go")

// Client enables managing resources in Fabric network.
type Client struct {
	provider          core.Providers
	identity          context.Identity
	discoveryProvider fab.DiscoveryProvider // used to get per channel discovery service(s)
	channelProvider   fab.ChannelProvider
	fabricProvider    fab.InfraProvider
	discovery         fab.DiscoveryService // global discovery service (detects all peers on the network)
	resource          api.Resource
	filter            TargetFilter
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
	core.Providers
	context.Identity
	DiscoveryProvider fab.DiscoveryProvider
	ChannelProvider   fab.ChannelProvider
	FabricProvider    fab.InfraProvider
}

type fabContext struct {
	core.Providers
	context.Identity
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure new
func WithDefaultTargetFilter(filter TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

// New returns a ResourceMgmtClient instance
func New(ctx Context, opts ...ClientOption) (*Client, error) {

	resource := resource.New(ctx)

	resourceClient := &Client{
		provider:          ctx,
		identity:          ctx,
		discoveryProvider: ctx.DiscoveryProvider,
		channelProvider:   ctx.ChannelProvider,
		fabricProvider:    ctx.FabricProvider,
		resource:          resource,
	}

	for _, opt := range opts {
		err := opt(resourceClient)
		if err != nil {
			return nil, err
		}
	}

	// setup global discovery service
	discovery, err := ctx.DiscoveryProvider.NewDiscoveryService("")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create global discovery service")
	}
	resourceClient.discovery = discovery
	//check if target filter was set - if not set the default
	if resourceClient.filter == nil {
		// Default target filter is based on user msp
		if ctx.MspID() == "" {
			return nil, errors.New("mspID not available in user context")
		}
		rcFilter := &MSPFilter{mspID: ctx.MspID()}
		resourceClient.filter = rcFilter
	}
	return resourceClient, nil
}

// JoinChannel allows for peers to join existing channel with optional custom options (specific peers, filtered peers)
func (rc *Client) JoinChannel(channelID string, options ...RequestOption) error {

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

	joinChannelRequest := api.JoinChannelRequest{
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

// helper method for calculating default targets
func (rc *Client) getDefaultTargets(discovery fab.DiscoveryService) ([]fab.Peer, error) {

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
func (rc *Client) calculateTargets(discovery fab.DiscoveryService, peers []fab.Peer, filter TargetFilter) ([]fab.Peer, error) {

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
func (rc *Client) isChaincodeInstalled(req InstallCCRequest, peer fab.Peer) (bool, error) {
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
func (rc *Client) InstallCC(req InstallCCRequest, options ...RequestOption) ([]InstallCCResponse, error) {

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

	responses := make([]InstallCCResponse, 0)
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
			response := InstallCCResponse{Target: target.URL(), Info: "already installed"}
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

	icr := api.InstallChaincodeRequest{Name: req.Name, Path: req.Path, Version: req.Version, Package: req.Package, Targets: peer.PeersToTxnProcessors(newTargets)}
	transactionProposalResponse, _, err := rc.resource.InstallChaincode(icr)
	for _, v := range transactionProposalResponse {
		logger.Debugf("Install chaincode '%s' endorser '%s' returned ProposalResponse status:%v", req.Name, v.Endorser, v.Status)

		response := InstallCCResponse{Target: v.Endorser, Status: v.Status}
		responses = append(responses, response)
	}

	if err != nil {
		return responses, errors.WithMessage(err, "InstallChaincode failed")
	}

	return responses, nil
}

func checkRequiredInstallCCParams(req InstallCCRequest) error {
	if req.Name == "" || req.Version == "" || req.Path == "" || req.Package == nil {
		return errors.New("Chaincode name, version, path and chaincode package are required")
	}
	return nil
}

// InstantiateCC instantiates chaincode using default settings
func (rc *Client) InstantiateCC(channelID string, req InstantiateCCRequest, options ...RequestOption) error {
	return rc.sendCCProposal(InstantiateChaincode, channelID, req, options...)
}

// UpgradeCC upgrades chaincode  with optional custom options (specific peers, filtered peers, timeout)
func (rc *Client) UpgradeCC(channelID string, req UpgradeCCRequest, options ...RequestOption) error {
	return rc.sendCCProposal(UpgradeChaincode, channelID, InstantiateCCRequest(req), options...)
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
// Returns the details of all chaincodes installed on a peer.
func (rc *Client) QueryInstalledChaincodes(proposalProcessor fab.ProposalProcessor) (*pb.ChaincodeQueryResponse, error) {
	return rc.resource.QueryInstalledChaincodes(proposalProcessor)
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on a peer for specific channel.
// Valid option is WithTarget. If not specified it will query any peer on this channel
func (rc *Client) QueryInstantiatedChaincodes(channelID string, options ...RequestOption) (*pb.ChaincodeQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	ctx := &fabContext{
		Providers: rc.provider,
		Identity:  rc.identity,
	}

	var target fab.ProposalProcessor
	if len(opts.Targets) >= 1 {
		target = opts.Targets[0]
	} else {
		// discover peers on this channel
		discovery, err := rc.discoveryProvider.NewDiscoveryService(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create channel discovery service")
		}
		// default filter will be applied (if any)
		targets, err := rc.getDefaultTargets(discovery)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get default target for query instantiated chaincodes")
		}

		// select random channel peer
		r := rand.New(rand.NewSource(time.Now().Unix()))
		randomNumber := r.Intn(len(targets))
		target = targets[randomNumber]
	}

	l, err := channel.NewLedger(ctx, channelID)
	if err != nil {
		return nil, err
	}

	responses, err := l.QueryInstantiatedChaincodes([]fab.ProposalProcessor{target})
	if err != nil {
		return nil, err
	}

	return responses[0], nil
}

// QueryChannels queries the names of all the channels that a peer has joined.
// Returns the details of all channels that peer has joined.
func (rc *Client) QueryChannels(proposalProcessor fab.ProposalProcessor) (*pb.ChannelQueryResponse, error) {
	return rc.resource.QueryChannels(proposalProcessor)
}

// sendCCProposal sends proposal for type  Instantiate, Upgrade
func (rc *Client) sendCCProposal(ccProposalType chaincodeProposalType, channelID string, req InstantiateCCRequest, options ...RequestOption) error {

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
	channelService, err := rc.channelProvider.ChannelService(rc.identity, channelID)
	if err != nil {
		return errors.WithMessage(err, "Unable to get channel service")
	}
	transactor, err := channelService.Transactor()
	if err != nil {
		return errors.WithMessage(err, "get channel transactor failed")
	}

	// create a transaction proposal for chaincode deployment
	deployProposal := chaincodeDeployRequest(req)
	deployCtx := fabContext{
		Providers: rc.provider,
		Identity:  rc.identity,
	}

	txid, err := txn.NewHeader(&deployCtx, channelID)
	if err != nil {
		return errors.WithMessage(err, "create transaction ID failed")
	}
	tp, err := createChaincodeDeployProposal(txid, ccProposalType, channelID, deployProposal)
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

	timeout := rc.provider.Config().TimeoutOrDefault(core.Execute)
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

func checkRequiredCCProposalParams(channelID string, req InstantiateCCRequest) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	if req.Name == "" || req.Version == "" || req.Path == "" || req.Policy == nil {
		return errors.New("Chaincode name, version, path and policy are required")
	}
	return nil
}

//prepareResmgmtOpts Reads Opts from Option array
func (rc *Client) prepareResmgmtOpts(options ...RequestOption) (Opts, error) {
	resmgmtOpts := Opts{}
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
func (rc *Client) SaveChannel(req SaveChannelRequest, options ...RequestOption) error {

	opts, err := rc.prepareRequestOpts(options...)
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
		Identity:  signer,
		Providers: rc.provider,
	}
	configSignature, err := resource.CreateConfigSignature(&sigCtx, chConfig)
	if err != nil {
		return errors.WithMessage(err, "signing configuration failed")
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	// Figure out orderer configuration
	var ordererCfg *core.OrdererConfig
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

	request := api.CreateChannelRequest{
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

//prepareRequestOpts prepares rrequest options
func (rc *Client) prepareRequestOpts(options ...RequestOption) (Opts, error) {
	opts := Opts{}
	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return opts, nil
}
