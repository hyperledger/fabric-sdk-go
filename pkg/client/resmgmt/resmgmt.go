/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt enables ability to update resources in a Fabric network.
package resmgmt

import (
	reqContext "context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/multi"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

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

//requestOptions contains options for operations performed by ResourceMgmtClient
type requestOptions struct {
	Targets       []fab.Peer                         // target peers
	TargetFilter  fab.TargetFilter                   // target filter
	Orderer       fab.Orderer                        // use specific orderer
	Timeouts      map[core.TimeoutType]time.Duration //timeout options for resmgmt operations
	ParentContext reqContext.Context                 //parent grpc context for resmgmt operations
}

//SaveChannelRequest used to save channel request
type SaveChannelRequest struct {
	ChannelID         string
	ChannelConfig     io.Reader             // ChannelConfig data source
	ChannelConfigPath string                // Convenience option to use the named file as ChannelConfig reader
	SigningIdentities []msp.SigningIdentity // Users that sign channel configuration
	// TODO: support pre-signed signature blocks
}

//RequestOption func for each Opts argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

var logger = logging.NewLogger("fabsdk/client")

// Client enables managing resources in Fabric network.
type Client struct {
	ctx       context.Client
	discovery fab.DiscoveryService // global discovery service (detects all peers on the network)
	filter    fab.TargetFilter
}

// mspFilter is default filter
type mspFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure new
func WithDefaultTargetFilter(filter fab.TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

// New returns a ResourceMgmtClient instance
func New(clientProvider context.ClientProvider, opts ...ClientOption) (*Client, error) {

	ctx, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create resmgmt client")
	}

	resourceClient := &Client{
		ctx: ctx,
	}

	for _, opt := range opts {
		err := opt(resourceClient)
		if err != nil {
			return nil, err
		}
	}

	// setup global discovery service
	discovery, err := ctx.DiscoveryProvider().CreateDiscoveryService("")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create global discovery service")
	}
	resourceClient.discovery = discovery
	//check if target filter was set - if not set the default
	if resourceClient.filter == nil {
		// Default target filter is based on user msp
		if ctx.Identifier().MSPID == "" {
			return nil, errors.New("mspID not available in user context")
		}
		rcFilter := &mspFilter{mspID: ctx.Identifier().MSPID}
		resourceClient.filter = rcFilter
	}
	return resourceClient, nil
}

// JoinChannel allows for peers to join existing channel with optional custom options (specific peers, filtered peers)
func (rc *Client) JoinChannel(channelID string, options ...RequestOption) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for JoinChannel")
	}

	//resolve timeouts
	rc.resolveTimeouts(&opts)

	//set parent request context for overall timeout
	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[core.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	targets, err := rc.calculateTargets(rc.discovery, opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for JoinChannel")
	}

	if len(targets) == 0 {
		return errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	orderer, err := rc.requestOrderer(&opts, channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to find orderer for request")
	}

	ordrReqCtx, ordrReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(core.OrdererResponse), contextImpl.WithParent(parentReqCtx))
	defer ordrReqCtxCancel()

	genesisBlock, err := resource.GenesisBlockFromOrderer(ordrReqCtx, channelID, orderer)
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	joinChannelRequest := api.JoinChannelRequest{
		GenesisBlock: genesisBlock,
	}

	peerReqCtx, peerReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(core.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer peerReqCtxCancel()
	err = resource.JoinChannel(peerReqCtx, joinChannelRequest, peersToTxnProcessors(targets))
	if err != nil {
		return errors.WithMessage(err, "join channel failed")
	}

	return nil
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
func (rc *Client) calculateTargets(discovery fab.DiscoveryService, peers []fab.Peer, filter fab.TargetFilter) ([]fab.Peer, error) {

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
func (rc *Client) isChaincodeInstalled(reqCtx reqContext.Context, req InstallCCRequest, peer fab.Peer) (bool, error) {

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, peer)
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

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for InstallCC")
	}

	//resolve timeouts
	rc.resolveTimeouts(&opts)

	//set parent request context for overall timeout
	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[core.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

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
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	responses := make([]InstallCCResponse, 0)
	var errs multi.Errors

	// Targets will be adjusted if cc has already been installed
	newTargets := make([]fab.Peer, 0)
	for _, target := range targets {
		reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(core.PeerResponse), contextImpl.WithParent(parentReqCtx))
		defer cancel()

		installed, err := rc.isChaincodeInstalled(reqCtx, req, target)
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

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(core.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	icr := api.InstallChaincodeRequest{Name: req.Name, Path: req.Path, Version: req.Version, Package: req.Package}
	transactionProposalResponse, _, err := resource.InstallChaincode(reqCtx, icr, peer.PeersToTxnProcessors(newTargets))
	for _, v := range transactionProposalResponse {
		logger.Debugf("Install chaincode '%s' endorser '%s' returned ProposalResponse status:%v", req.Name, v.Endorser, v.Status)

		response := InstallCCResponse{Target: v.Endorser, Status: v.Status}
		responses = append(responses, response)
	}

	if err != nil {
		return responses, errors.WithMessage(err, "InstallChaincode failed")
	}
	if len(errs) > 0 {
		return responses, errs
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

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for InstantiateCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.ResMgmt)
	defer cancel()

	return rc.sendCCProposal(reqCtx, InstantiateChaincode, channelID, req, opts)
}

// UpgradeCC upgrades chaincode  with optional custom options (specific peers, filtered peers, timeout)
func (rc *Client) UpgradeCC(channelID string, req UpgradeCCRequest, options ...RequestOption) error {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for UpgradeCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.ResMgmt)
	defer cancel()

	return rc.sendCCProposal(reqCtx, UpgradeChaincode, channelID, InstantiateCCRequest(req), opts)
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
// Returns the details of all chaincodes installed on a peer.
func (rc *Client) QueryInstalledChaincodes(options ...RequestOption) (*pb.ChaincodeQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.PeerResponse)
	defer cancel()

	return resource.QueryInstalledChaincodes(reqCtx, opts.Targets[0])
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on a peer for specific channel.
// Valid option is WithTarget. If not specified it will query any peer on this channel
func (rc *Client) QueryInstantiatedChaincodes(channelID string, options ...RequestOption) (*pb.ChaincodeQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	var target fab.ProposalProcessor
	if len(opts.Targets) >= 1 {
		target = opts.Targets[0]
	} else {
		// discover peers on this channel
		discovery, err := rc.ctx.DiscoveryProvider().CreateDiscoveryService(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create channel discovery service")
		}
		// default filter will be applied (if any)
		targets, err := rc.getDefaultTargets(discovery)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get default target for query instantiated chaincodes")
		}

		// select random channel peer
		randomNumber := rand.Intn(len(targets))
		target = targets[randomNumber]
	}

	l, err := channel.NewLedger(channelID)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.PeerResponse)
	defer cancel()

	// Channel service membership is required to verify signature
	channelService, err := rc.ctx.ChannelProvider().ChannelService(rc.ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to get channel service")
	}

	membership, err := channelService.Membership()
	if err != nil {
		return nil, errors.WithMessage(err, "membership creation failed")
	}

	responses, err := l.QueryInstantiatedChaincodes(reqCtx, []fab.ProposalProcessor{target}, &verifier.Signature{Membership: membership})
	if err != nil {
		return nil, err
	}

	return responses[0], nil
}

// QueryChannels queries the names of all the channels that a peer has joined.
// Returns the details of all channels that peer has joined.
func (rc *Client) QueryChannels(options ...RequestOption) (*pb.ChannelQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.PeerResponse)
	defer cancel()

	return resource.QueryChannels(reqCtx, opts.Targets[0])

}

// sendCCProposal sends proposal for type  Instantiate, Upgrade
func (rc *Client) sendCCProposal(reqCtx reqContext.Context, ccProposalType chaincodeProposalType, channelID string, req InstantiateCCRequest, opts requestOptions) error {

	if err := checkRequiredCCProposalParams(channelID, req); err != nil {
		return err
	}

	// per channel discovery service
	discovery, err := rc.ctx.DiscoveryProvider().CreateDiscoveryService(channelID)
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
		return errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	// Get transactor on the channel to create and send the deploy proposal
	channelService, err := rc.ctx.ChannelProvider().ChannelService(rc.ctx, channelID)
	if err != nil {
		return errors.WithMessage(err, "Unable to get channel service")
	}

	chConfig, err := channelService.ChannelConfig()
	if err != nil {
		return errors.WithMessage(err, "get channel config failed")
	}
	transactor, err := rc.ctx.InfraProvider().CreateChannelTransactor(reqCtx, chConfig)
	if err != nil {
		return errors.WithMessage(err, "get channel transactor failed")
	}

	// create a transaction proposal for chaincode deployment
	deployProposal := chaincodeDeployRequest(req)

	txid, err := txn.NewHeader(rc.ctx, channelID)
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

	// Membership is required to verify signature
	membership, err := channelService.Membership()
	if err != nil {
		return errors.WithMessage(err, "membership creation failed")
	}

	// Verify signature(s)
	sv := &verifier.Signature{Membership: membership}
	for _, r := range txProposalResponse {
		if err := sv.Verify(r); err != nil {
			return errors.WithMessage(err, "Failed to verify signature")
		}
	}

	eventService, err := channelService.EventService()
	if err != nil {
		return errors.WithMessage(err, "unable to get event service")
	}

	// Register for commit event
	reg, statusNotifier, err := eventService.RegisterTxStatusEvent(string(tp.TxnID))
	if err != nil {
		return errors.WithMessage(err, "error registering for TxStatus event")
	}
	defer eventService.Unregister(reg)

	transactionRequest := fab.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: txProposalResponse,
	}
	if _, err = createAndSendTransaction(transactor, transactionRequest); err != nil {
		return errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	select {
	case txStatus := <-statusNotifier:
		if txStatus.TxValidationCode == pb.TxValidationCode_VALID {
			return nil
		}
		return status.New(status.EventServerStatus, int32(txStatus.TxValidationCode), "instantiateOrUpgradeCC failed", nil)
	case <-reqCtx.Done():
		return errors.New("instantiateOrUpgradeCC timed out or cancelled")
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

	if req.ChannelConfigPath != "" {
		configReader, err := os.Open(req.ChannelConfigPath)
		if err != nil {
			return errors.Wrapf(err, "opening channel config file failed")
		}
		defer configReader.Close()
		req.ChannelConfig = configReader
	}

	if req.ChannelID == "" || req.ChannelConfig == nil {
		return errors.New("must provide channel ID and channel config")
	}

	logger.Debugf("saving channel: %s", req.ChannelID)

	// Signing user has to belong to one of configured channel organisations
	// In case that order org is one of channel orgs we can use context user
	var signers []msp.SigningIdentity

	if len(req.SigningIdentities) > 0 {
		for _, id := range req.SigningIdentities {
			if id != nil {
				signers = append(signers, id)
			}
		}
	} else if rc.ctx != nil {
		signers = append(signers, rc.ctx)
	} else {
		return errors.New("must provide signing user")
	}

	configTx, err := ioutil.ReadAll(req.ChannelConfig)
	if err != nil {
		return errors.WithMessage(err, "reading channel config file failed")
	}

	chConfig, err := resource.ExtractChannelConfig(configTx)
	if err != nil {
		return errors.WithMessage(err, "extracting channel config failed")
	}

	var configSignatures []*common.ConfigSignature
	for _, signer := range signers {

		sigCtx := contextImpl.Client{
			SigningIdentity: signer,
			Providers:       rc.ctx,
		}

		configSignature, err := resource.CreateConfigSignature(&sigCtx, chConfig)
		if err != nil {
			return errors.WithMessage(err, "signing configuration failed")
		}
		configSignatures = append(configSignatures, configSignature)
	}

	orderer, err := rc.requestOrderer(&opts, req.ChannelID)
	if err != nil {
		return errors.WithMessage(err, "failed to find orderer for request")
	}

	request := api.CreateChannelRequest{
		Name:       req.ChannelID,
		Orderer:    orderer,
		Config:     chConfig,
		Signatures: configSignatures,
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.OrdererResponse)
	defer cancel()

	_, err = resource.CreateChannel(reqCtx, request)
	if err != nil {
		return errors.WithMessage(err, "create channel failed")
	}

	return nil
}

// QueryConfigFromOrderer config returns channel configuration from orderer
// Valid request option is WithOrdererID
// If orderer id is not provided orderer will be defaulted to channel orderer (if configured) or random orderer from config
func (rc *Client) QueryConfigFromOrderer(channelID string, options ...RequestOption) (fab.ChannelCfg, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	orderer, err := rc.requestOrderer(&opts, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to find orderer for request")
	}

	channelConfig, err := chconfig.New(channelID, chconfig.WithOrderer(orderer))
	if err != nil {
		return nil, errors.WithMessage(err, "QueryConfig failed")
	}

	reqCtx, cancel := rc.createRequestContext(opts, core.OrdererResponse)
	defer cancel()

	return channelConfig.Query(reqCtx)

}

func (rc *Client) requestOrderer(opts *requestOptions, channelID string) (fab.Orderer, error) {
	if opts.Orderer != nil {
		return opts.Orderer, nil
	}

	ordererCfg, err := rc.ordererConfig(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "orderer not found")
	}

	orderer, err := rc.ctx.InfraProvider().CreateOrdererFromConfig(ordererCfg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create orderer from config")
	}
	return orderer, nil

}

func (rc *Client) ordererConfig(channelID string) (*core.OrdererConfig, error) {
	orderers, err := rc.ctx.Config().ChannelOrderers(channelID)

	// TODO: Not sure that we should fallback to global orderers section.
	// For now - not doing so.
	//if err != nil || len(orderers) == 0 {
	//	orderers, err = rc.ctx.Config().OrderersConfig()
	//}

	if err != nil {
		return nil, errors.WithMessage(err, "orderers lookup failed")
	}
	if len(orderers) == 0 {
		return nil, errors.New("no orderers found")
	}

	// random channel orderer
	randomNumber := rand.Intn(len(orderers))
	return &orderers[randomNumber], nil
}

// prepareRequestOpts prepares request options
func (rc *Client) prepareRequestOpts(options ...RequestOption) (requestOptions, error) {
	opts := requestOptions{}
	for _, option := range options {
		err := option(rc.ctx, &opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return opts, nil
}

//createRequestContext creates request context for grpc
func (rc *Client) createRequestContext(opts requestOptions, defaultTimeoutType core.TimeoutType) (reqContext.Context, reqContext.CancelFunc) {

	rc.resolveTimeouts(&opts)

	if opts.Timeouts[defaultTimeoutType] == 0 {
		opts.Timeouts[defaultTimeoutType] = rc.ctx.Config().TimeoutOrDefault(defaultTimeoutType)
	}

	return contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[defaultTimeoutType]), contextImpl.WithParent(opts.ParentContext))
}

//resolveTimeouts sets default for timeouts from config if not provided through opts
func (rc *Client) resolveTimeouts(opts *requestOptions) {

	if opts.Timeouts == nil {
		opts.Timeouts = make(map[core.TimeoutType]time.Duration)
	}

	if opts.Timeouts[core.ResMgmt] == 0 {
		opts.Timeouts[core.ResMgmt] = rc.ctx.Config().TimeoutOrDefault(core.ResMgmt)
	}

	if opts.Timeouts[core.OrdererResponse] == 0 {
		opts.Timeouts[core.OrdererResponse] = rc.ctx.Config().TimeoutOrDefault(core.OrdererResponse)
	}

	if opts.Timeouts[core.PeerResponse] == 0 {
		opts.Timeouts[core.PeerResponse] = rc.ctx.Config().TimeoutOrDefault(core.PeerResponse)
	}
}
