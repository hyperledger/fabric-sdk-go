/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resmgmt enables creation and update of resources on a Fabric network.
// It allows administrators to create and/or update channnels, and for peers to join channels.
// Administrators can also perform chaincode related operations on a peer, such as
// installing, instantiating, and upgrading chaincode.
//
//  Basic Flow:
//  1) Prepare client context
//  2) Create resource managememt client
//  3) Create new channel
//  4) Peer(s) join channel
//  5) Install chaincode onto peer(s) filesystem
//  6) Instantiate chaincode on channel
//  7) Query peer for channels, installed/instantiated chaincodes etc.
package resmgmt

import (
	reqContext "context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// InstallCCRequest contains install chaincode request parameters
type InstallCCRequest struct {
	Name    string
	Path    string
	Version string
	Package *resource.CCPackage
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

// InstantiateCCResponse contains response parameters for instantiate chaincode
type InstantiateCCResponse struct {
	TransactionID fab.TransactionID
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

// UpgradeCCResponse contains response parameters for upgrade chaincode
type UpgradeCCResponse struct {
	TransactionID fab.TransactionID
}

//requestOptions contains options for operations performed by ResourceMgmtClient
type requestOptions struct {
	Targets       []fab.Peer                        // target peers
	TargetFilter  fab.TargetFilter                  // target filter
	Orderer       fab.Orderer                       // use specific orderer
	Timeouts      map[fab.TimeoutType]time.Duration //timeout options for resmgmt operations
	ParentContext reqContext.Context                //parent grpc context for resmgmt operations
	Retry         retry.Opts
}

//SaveChannelRequest holds parameters for save channel request
type SaveChannelRequest struct {
	ChannelID         string
	ChannelConfig     io.Reader             // ChannelConfig data source
	ChannelConfigPath string                // Convenience option to use the named file as ChannelConfig reader
	SigningIdentities []msp.SigningIdentity // Users that sign channel configuration
	// TODO: support pre-signed signature blocks
}

// SaveChannelResponse contains response parameters for save channel
type SaveChannelResponse struct {
	TransactionID fab.TransactionID
}

//RequestOption func for each Opts argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

var logger = logging.NewLogger("fabsdk/client")

// Client enables managing resources in Fabric network.
type Client struct {
	ctx              context.Client
	filter           fab.TargetFilter
	localCtxProvider context.LocalProvider
}

// mspFilter filters peers by MSP ID
type mspFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure default target filter per client
func WithDefaultTargetFilter(filter fab.TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

// New returns a resource management client instance.
func New(ctxProvider context.ClientProvider, opts ...ClientOption) (*Client, error) {

	ctx, err := ctxProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create resmgmt client due to context error")
	}

	if ctx.Identifier().MSPID == "" {
		return nil, errors.New("mspID not available in user context")
	}

	resourceClient := &Client{
		ctx: ctx,
	}

	for _, opt := range opts {
		err1 := opt(resourceClient)
		if err1 != nil {
			return nil, err1
		}
	}

	if resourceClient.localCtxProvider == nil {
		resourceClient.localCtxProvider = func() (context.Local, error) {
			return contextImpl.NewLocal(
				func() (context.Client, error) {
					return resourceClient.ctx, nil
				},
			)
		}
	}

	return resourceClient, nil
}

// JoinChannel allows for peers to join existing channel with optional custom options (specific peers, filtered peers). If peer(s) are not specified in options it will default to all peers that belong to client's MSP.
//  Parameters:
//  channel is manadatory channel name
//  options holds optional request options
//
//  Returns:
//  an error if join fails
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
	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	targets, err := rc.calculateTargets(opts.Targets, opts.TargetFilter)
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

	ordrReqCtx, ordrReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.OrdererResponse), contextImpl.WithParent(parentReqCtx))
	defer ordrReqCtxCancel()

	genesisBlock, err := resource.GenesisBlockFromOrderer(ordrReqCtx, channelID, orderer, resource.WithRetry(opts.Retry))
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	joinChannelRequest := resource.JoinChannelRequest{
		GenesisBlock: genesisBlock,
	}

	peerReqCtx, peerReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer peerReqCtxCancel()
	err = resource.JoinChannel(peerReqCtx, joinChannelRequest, peersToTxnProcessors(targets), resource.WithRetry(opts.Retry))
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

func (rc *Client) resolveDefaultTargets(opts *requestOptions) ([]fab.Peer, error) {
	if len(opts.Targets) != 0 {
		return opts.Targets, nil
	}

	localCtx, err := rc.localCtxProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create local context")
	}

	targets, err := rc.getDefaultTargets(localCtx.LocalDiscoveryService())
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.WithMessage(err, "no local targets for InstallCC")
	}

	return targets, nil
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
func (rc *Client) calculateTargets(targets []fab.Peer, filter fab.TargetFilter) ([]fab.Peer, error) {

	targetFilter := filter

	if len(targets) == 0 {
		// Retrieve targets from discovery
		localCtx, err := rc.localCtxProvider()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create local context")
		}
		targets, err = localCtx.LocalDiscoveryService().GetPeers()
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
func (rc *Client) isChaincodeInstalled(reqCtx reqContext.Context, req InstallCCRequest, peer fab.ProposalProcessor, retryOpts retry.Opts) (bool, error) {

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, peer, resource.WithRetry(retryOpts))
	if err != nil {
		return false, err
	}

	logger.Debugf("isChaincodeInstalled: %+v", chaincodeQueryResponse)

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == req.Name && chaincode.Version == req.Version && chaincode.Path == req.Path {
			return true, nil
		}
	}

	return false, nil
}

// InstallCC allows administrators to install chaincode onto the filesystem of a peer.
// If peer(s) are not specified in options it will default to all peers that belong to admin's MSP.
//  Parameters:
//  req holds info about mandatory chaincode name, path, version and policy
//  options holds optional request options
//
//  Returns:
//  install chaincode proposal responses from peer(s)
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
	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	//Default targets when targets are not provided in options
	defaultTargets, err := rc.resolveDefaultTargets(&opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get default targets for InstallCC")
	}

	targets, err := rc.calculateTargets(defaultTargets, opts.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for install cc")
	}

	if len(targets) == 0 {
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	responses, newTargets, errs := rc.adjustTargets(targets, req, opts.Retry, parentReqCtx)

	if len(newTargets) == 0 {
		// CC is already installed on all targets and/or
		// we are unable to verify if cc is installed on target(s)
		return responses, errs.ToError()
	}

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	responses = rc.sendIntallCCRequest(req, reqCtx, newTargets, responses)

	if err != nil {
		installErrs, ok := err.(multi.Errors)
		if ok {
			errs = append(errs, installErrs)
		} else {
			errs = append(errs, err)
		}
	}

	return responses, errs.ToError()
}

func (rc *Client) sendIntallCCRequest(req InstallCCRequest, reqCtx reqContext.Context, newTargets []fab.Peer, responses []InstallCCResponse) []InstallCCResponse {
	icr := resource.InstallChaincodeRequest{Name: req.Name, Path: req.Path, Version: req.Version, Package: req.Package}
	transactionProposalResponse, _, _ := resource.InstallChaincode(reqCtx, icr, peer.PeersToTxnProcessors(newTargets))
	for _, v := range transactionProposalResponse {
		logger.Debugf("Install chaincode '%s' endorser '%s' returned ProposalResponse status:%v", req.Name, v.Endorser, v.Status)

		response := InstallCCResponse{Target: v.Endorser, Status: v.Status}
		responses = append(responses, response)
	}
	return responses
}

func (rc *Client) adjustTargets(targets []fab.Peer, req InstallCCRequest, retry retry.Opts, parentReqCtx reqContext.Context) ([]InstallCCResponse, []fab.Peer, multi.Errors) {
	errs := multi.Errors{}

	responses := make([]InstallCCResponse, 0)

	// Targets will be adjusted if cc has already been installed
	newTargets := make([]fab.Peer, 0)
	for _, target := range targets {
		reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.PeerResponse), contextImpl.WithParent(parentReqCtx))
		defer cancel()

		installed, err1 := rc.isChaincodeInstalled(reqCtx, req, target, retry)
		if err1 != nil {
			// Add to errors with unable to verify error message
			errs = append(errs, errors.Errorf("unable to verify if cc is installed on %s. Got error: %s", target.URL(), err1))
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

	return responses, newTargets, errs

}

func checkRequiredInstallCCParams(req InstallCCRequest) error {
	if req.Name == "" || req.Version == "" || req.Path == "" || req.Package == nil {
		return errors.New("Chaincode name, version, path and chaincode package are required")
	}
	return nil
}

// InstantiateCC instantiates chaincode with optional custom options (specific peers, filtered peers, timeout). If peer(s) are not specified
// in options it will default to all channel peers.
//  Parameters:
//  channel is manadatory channel name
//  req holds info about mandatory chaincode name, path, version and policy
//  options holds optional request options
//
//  Returns:
//  instantiate chaincode response with transaction ID
func (rc *Client) InstantiateCC(channelID string, req InstantiateCCRequest, options ...RequestOption) (InstantiateCCResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return InstantiateCCResponse{}, errors.WithMessage(err, "failed to get opts for InstantiateCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	txID, err := rc.sendCCProposal(reqCtx, InstantiateChaincode, channelID, req, opts)
	return InstantiateCCResponse{TransactionID: txID}, err
}

// UpgradeCC upgrades chaincode with optional custom options (specific peers, filtered peers, timeout). If peer(s) are not specified in options
// it will default to all channel peers.
//  Parameters:
//  channel is manadatory channel name
//  req holds info about mandatory chaincode name, path, version and policy
//  options holds optional request options
//
//  Returns:
//  upgrade chaincode response with transaction ID
func (rc *Client) UpgradeCC(channelID string, req UpgradeCCRequest, options ...RequestOption) (UpgradeCCResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return UpgradeCCResponse{}, errors.WithMessage(err, "failed to get opts for UpgradeCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	txID, err := rc.sendCCProposal(reqCtx, UpgradeChaincode, channelID, InstantiateCCRequest(req), opts)
	return UpgradeCCResponse{TransactionID: txID}, err
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
//  Parameters:
//  options hold optional request options
//  Note: One target(peer) has to be specified using either WithTargetURLs or WithTargets request option
//
//  Returns:
//  list of installed chaincodes on specified peer
func (rc *Client) QueryInstalledChaincodes(options ...RequestOption) (*pb.ChaincodeQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.PeerResponse)
	defer cancel()

	return resource.QueryInstalledChaincodes(reqCtx, opts.Targets[0], resource.WithRetry(opts.Retry))
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on a peer for specific channel. If peer is not specified in options it will query random peer on this channel.
//  Parameters:
//  channel is manadatory channel name
//  options hold optional request options
//
//  Returns:
//  list of instantiated chaincodes
func (rc *Client) QueryInstantiatedChaincodes(channelID string, options ...RequestOption) (*pb.ChaincodeQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	chCtx, err := contextImpl.NewChannel(
		func() (context.Client, error) {
			return rc.ctx, nil
		},
		channelID,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	var target fab.ProposalProcessor
	if len(opts.Targets) >= 1 {
		target = opts.Targets[0]
	} else {
		// discover peers on this channel
		discovery, err := chCtx.ChannelService().Discovery()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get discovery service")
		}
		// default filter will be applied (if any)
		targets, err2 := rc.getDefaultTargets(discovery)
		if err2 != nil {
			return nil, errors.WithMessage(err2, "failed to get default target for query instantiated chaincodes")
		}

		// Filter by MSP since the LSCC only allows local calls
		targets = filterTargets(targets, &mspFilter{mspID: chCtx.Identifier().MSPID})

		if len(targets) == 0 {
			return nil, errors.Errorf("no targets in MSP [%s]", chCtx.Identifier().MSPID)
		}

		// select random channel peer
		randomNumber := rand.Intn(len(targets))
		target = targets[randomNumber]
	}

	l, err := channel.NewLedger(channelID)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.PeerResponse)
	defer cancel()

	// Channel service membership is required to verify signature
	channelService := chCtx.ChannelService()

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
//  Parameters:
//  options hold optional request options
//  Note: One target(peer) has to be specified using either WithTargetURLs or WithTargets request option
//
//  Returns:
//  all channels that peer has joined
func (rc *Client) QueryChannels(options ...RequestOption) (*pb.ChannelQueryResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, err
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.PeerResponse)
	defer cancel()

	return resource.QueryChannels(reqCtx, opts.Targets[0], resource.WithRetry(opts.Retry))

}

// validateSendCCProposal
func (rc *Client) getCCProposalTargets(channelID string, req InstantiateCCRequest, opts requestOptions) ([]fab.Peer, error) {

	chCtx, err := contextImpl.NewChannel(
		func() (context.Client, error) {
			return rc.ctx, nil
		},
		channelID,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	// per channel discovery service
	discovery, err := chCtx.ChannelService().Discovery()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get discovery service")
	}

	//Default targets when targets are not provided in options
	if len(opts.Targets) == 0 {
		opts.Targets, err = rc.getDefaultTargets(discovery)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get default targets for cc proposal")
		}
	}

	targets, err := rc.calculateTargets(opts.Targets, opts.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for cc proposal")
	}

	if len(targets) == 0 {
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}
	return targets, nil
}

// createTP
func (rc *Client) createTP(req InstantiateCCRequest, channelID string, ccProposalType chaincodeProposalType) (*fab.TransactionProposal, fab.TransactionID, error) {
	deployProposal := chaincodeDeployRequest(req)

	txID, err := txn.NewHeader(rc.ctx, channelID)
	if err != nil {
		return nil, fab.EmptyTransactionID, errors.WithMessage(err, "create transaction ID failed")
	}

	tp, err := createChaincodeDeployProposal(txID, ccProposalType, channelID, deployProposal)
	if err != nil {
		return nil, txID.TransactionID(), errors.WithMessage(err, "creating chaincode deploy transaction proposal failed")
	}
	return tp, txID.TransactionID(), nil
}

func (rc *Client) verifyTPSignature(channelService fab.ChannelService, txProposalResponse []*fab.TransactionProposalResponse) error {
	// Membership is required to verify signature
	membership, err := channelService.Membership()
	if err != nil {
		return errors.WithMessage(err, "membership creation failed")
	}

	sv := &verifier.Signature{Membership: membership}
	for _, r := range txProposalResponse {
		if err := sv.Verify(r); err != nil {
			return errors.WithMessage(err, "Failed to verify signature")
		}
	}
	return nil
}

// sendCCProposal sends proposal for type  Instantiate, Upgrade
func (rc *Client) sendCCProposal(reqCtx reqContext.Context, ccProposalType chaincodeProposalType, channelID string, req InstantiateCCRequest, opts requestOptions) (fab.TransactionID, error) {
	if err := checkRequiredCCProposalParams(channelID, req); err != nil {
		return fab.EmptyTransactionID, err
	}

	targets, err := rc.getCCProposalTargets(channelID, req, opts)
	if err != nil {
		return fab.EmptyTransactionID, err
	}
	// Get transactor on the channel to create and send the deploy proposal
	channelService, err := rc.ctx.ChannelProvider().ChannelService(rc.ctx, channelID)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "Unable to get channel service")
	}

	transactor, err := channelService.Transactor(reqCtx)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "get channel transactor failed")
	}

	// create a transaction proposal for chaincode deployment
	tp, txnID, err := rc.createTP(req, channelID, ccProposalType)
	if err != nil {
		return txnID, err
	}

	// Process and send transaction proposal
	txProposalResponse, err := transactor.SendTransactionProposal(tp, peersToTxnProcessors(targets))
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending deploy transaction proposal failed")
	}

	// Verify signature(s)
	err = rc.verifyTPSignature(channelService, txProposalResponse)
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "sending deploy transaction proposal failed to verify signature")
	}

	eventService, err := channelService.EventService()
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "unable to get event service")
	}

	// send transaction and check event
	return rc.sendTransactionAndCheckEvent(eventService, tp, txProposalResponse, transactor, reqCtx)

}

func (rc *Client) sendTransactionAndCheckEvent(eventService fab.EventService, tp *fab.TransactionProposal, txProposalResponse []*fab.TransactionProposalResponse,
	transac fab.Transactor, reqCtx reqContext.Context) (fab.TransactionID, error) {
	// Register for commit event
	reg, statusNotifier, err := eventService.RegisterTxStatusEvent(string(tp.TxnID))
	if err != nil {
		return tp.TxnID, errors.WithMessage(err, "error registering for TxStatus event")
	}
	defer eventService.Unregister(reg)

	transactionRequest := fab.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: txProposalResponse,
	}
	if _, err := createAndSendTransaction(transac, transactionRequest); err != nil {
		return tp.TxnID, errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	select {
	case txStatus := <-statusNotifier:
		if txStatus.TxValidationCode == pb.TxValidationCode_VALID {
			return fab.TransactionID(txStatus.TxID), nil
		}
		return fab.TransactionID(txStatus.TxID), status.New(status.EventServerStatus, int32(txStatus.TxValidationCode), "instantiateOrUpgradeCC failed", nil)
	case <-reqCtx.Done():
		return tp.TxnID, errors.New("instantiateOrUpgradeCC timed out or cancelled")
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

// SaveChannel creates or updates channel.
//  Parameters:
//  req holds info about mandatory channel name and configuration
//  options holds optional request options
//
//  Returns:
//  save channel response with transaction ID
func (rc *Client) SaveChannel(req SaveChannelRequest, options ...RequestOption) (SaveChannelResponse, error) {

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return SaveChannelResponse{}, err
	}

	if req.ChannelConfigPath != "" {
		configReader, err1 := os.Open(req.ChannelConfigPath)
		if err1 != nil {
			return SaveChannelResponse{}, errors.Wrapf(err1, "opening channel config file failed")
		}
		defer loggedClose(configReader)
		req.ChannelConfig = configReader
	}

	err = rc.validateSaveChannelRequest(req)
	if err != nil {
		return SaveChannelResponse{}, errors.WithMessage(err, "reading channel config file failed")
	}

	logger.Debugf("saving channel: %s", req.ChannelID)

	configTx, err := ioutil.ReadAll(req.ChannelConfig)
	if err != nil {
		return SaveChannelResponse{}, errors.WithMessage(err, "reading channel config file failed")
	}

	chConfig, err := resource.ExtractChannelConfig(configTx)
	if err != nil {
		return SaveChannelResponse{}, errors.WithMessage(err, "extracting channel config failed")
	}

	orderer, err := rc.requestOrderer(&opts, req.ChannelID)
	if err != nil {
		return SaveChannelResponse{}, errors.WithMessage(err, "failed to find orderer for request")
	}

	configSignatures, err := rc.getConfigSignatures(req, chConfig)
	if err != nil {
		return SaveChannelResponse{}, err
	}

	request := resource.CreateChannelRequest{
		Name:       req.ChannelID,
		Orderer:    orderer,
		Config:     chConfig,
		Signatures: configSignatures,
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.OrdererResponse)
	defer cancel()

	txID, err := resource.CreateChannel(reqCtx, request, resource.WithRetry(opts.Retry))
	if err != nil {
		return SaveChannelResponse{}, errors.WithMessage(err, "create channel failed")
	}

	return SaveChannelResponse{TransactionID: txID}, nil
}

func (rc *Client) validateSaveChannelRequest(req SaveChannelRequest) error {

	if req.ChannelID == "" || req.ChannelConfig == nil {
		return errors.New("must provide channel ID and channel config")
	}
	return nil
}

func (rc *Client) getConfigSignatures(req SaveChannelRequest, chConfig []byte) ([]*common.ConfigSignature, error) {

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
		return nil, errors.New("must provide signing user")
	}

	var configSignatures []*common.ConfigSignature
	for _, signer := range signers {

		sigCtx := contextImpl.Client{
			SigningIdentity: signer,
			Providers:       rc.ctx,
		}

		configSignature, err1 := resource.CreateConfigSignature(&sigCtx, chConfig)
		if err1 != nil {
			return nil, errors.WithMessage(err1, "signing configuration failed")
		}
		configSignatures = append(configSignatures, configSignature)
	}

	return configSignatures, nil

}

func loggedClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		logger.Warnf("closing resource failed: %s", err)
	}
}

// QueryConfigFromOrderer config returns channel configuration from orderer. If orderer is not provided using options it will be defaulted to channel orderer (if configured) or random orderer from configuration.
//  Parameters:
//  channelID is mandatory channel ID
//  options holds optional request options
//
//  Returns:
//  channel configuration
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

	reqCtx, cancel := rc.createRequestContext(opts, fab.OrdererResponse)
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

func (rc *Client) ordererConfig(channelID string) (*fab.OrdererConfig, error) {
	orderers, ok := rc.ctx.EndpointConfig().ChannelOrderers(channelID)

	// TODO: Not sure that we should fallback to global orderers section.
	// For now - not doing so.
	//if err != nil || len(orderers) == 0 {
	//	orderers, err = rc.ctx.Config().OrderersConfig()
	//}

	if !ok {
		return nil, errors.New("orderers lookup failed")
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
			return opts, errors.WithMessage(err, "failed to read opts in resmgmt")
		}
	}

	if len(opts.Targets) > 0 && opts.TargetFilter != nil {
		return opts, errors.New("If targets are provided, filter cannot be provided")
	}

	return opts, nil
}

//createRequestContext creates request context for grpc
func (rc *Client) createRequestContext(opts requestOptions, defaultTimeoutType fab.TimeoutType) (reqContext.Context, reqContext.CancelFunc) {

	rc.resolveTimeouts(&opts)

	if opts.Timeouts[defaultTimeoutType] == 0 {
		opts.Timeouts[defaultTimeoutType] = rc.ctx.EndpointConfig().Timeout(defaultTimeoutType)
	}

	return contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[defaultTimeoutType]), contextImpl.WithParent(opts.ParentContext))
}

//resolveTimeouts sets default for timeouts from config if not provided through opts
func (rc *Client) resolveTimeouts(opts *requestOptions) {

	if opts.Timeouts == nil {
		opts.Timeouts = make(map[fab.TimeoutType]time.Duration)
	}

	if opts.Timeouts[fab.ResMgmt] == 0 {
		opts.Timeouts[fab.ResMgmt] = rc.ctx.EndpointConfig().Timeout(fab.ResMgmt)
	}

	if opts.Timeouts[fab.OrdererResponse] == 0 {
		opts.Timeouts[fab.OrdererResponse] = rc.ctx.EndpointConfig().Timeout(fab.OrdererResponse)
	}

	if opts.Timeouts[fab.PeerResponse] == 0 {
		opts.Timeouts[fab.PeerResponse] = rc.ctx.EndpointConfig().Timeout(fab.PeerResponse)
	}
}
