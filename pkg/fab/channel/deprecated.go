// +build deprecated

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

const (
	lsccDeploy  = "deploy"
	lsccUpgrade = "upgrade"
	escc        = "escc"
	vscc        = "vscc"

	InstantiateChaincode ChaincodeProposalType = iota
	UpgradeChaincode
)

// ChaincodeProposalType reflects transitions in the chaincode lifecycle
type ChaincodeProposalType int

// ChaincodeDeployRequest holds parameters for creating an instantiate or upgrade chaincode proposal.
type ChaincodeDeployRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

// Channel  captures settings for a channel, which is created by
// the orderers to isolate transactions delivery to peers participating on channel.
type Channel struct {
	name          string // aka channel ID
	peers         map[string]fab.Peer
	orderers      map[string]fab.Orderer
	clientContext context.Client
	primaryPeer   fab.Peer
	mspManager    msp.MSPManager
	anchorPeers   []*fab.OrgAnchorPeer
	transactor    fab.Transactor
	initialized   bool
}

// New represents a channel in a Fabric network.
// name: used to identify different channel instances. The naming of channel instances
// is enforced by the ordering service and must be unique within the blockchain network.
// client: Provides operational context such as submitting User etc.
func New(ctx context.Client, cfg fab.ChannelCfg) (*Channel, error) {
	if ctx == nil {
		return nil, errors.Errorf("client is required")
	}
	p := make(map[string]fab.Peer)
	o := make(map[string]fab.Orderer)

	transactor, err := NewTransactor(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "transactor creation failed")
	}

	c := Channel{
		name:          cfg.Name(),
		peers:         p,
		orderers:      o,
		clientContext: ctx,
		transactor:    transactor,
	}

	mspManager := msp.NewMSPManager()
	if len(cfg.Msps()) > 0 {
		msps, err := loadMSPs(cfg.Msps(), ctx.CryptoSuite())
		if err != nil {
			return nil, errors.WithMessage(err, "load MSPs from config failed")
		}

		if err := mspManager.Setup(msps); err != nil {
			return nil, errors.WithMessage(err, "MSPManager Setup failed")
		}

		for _, msp := range msps {
			for _, cert := range msp.GetTLSRootCerts() {
				addCertsToConfig(ctx.Config(), cert)
			}

			for _, cert := range msp.GetTLSIntermediateCerts() {
				addCertsToConfig(ctx.Config(), cert)
			}
		}
	}

	c.mspManager = mspManager
	c.anchorPeers = cfg.AnchorPeers()

	// Add orderer if specified in config
	for _, name := range cfg.Orderers() {
		//Get orderer config by orderer address
		oCfg, err := getOrdererConfig(ctx.Config(), name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve orderer config...: %s", err)
		}

		var o *orderer.Orderer
		if oCfg == nil {
			o, err = orderer.New(ctx.Config(), orderer.WithURL(name), orderer.WithServerName(resolveOrdererAddress(name)))
		} else {
			o, err = orderer.New(ctx.Config(), orderer.FromOrdererConfig(oCfg))
		}

		if err != nil {
			return nil, errors.WithMessage(err, "failed to create new orderer from config")
		}

		c.orderers[o.URL()] = o
	}

	logger.Debugf("Constructed channel instance for channel %s: %v", c.name, c)

	return &c, nil
}

// Name returns the channel name.
func (c *Channel) Name() string {
	return c.name
}

// AddPeer adds a peer endpoint to channel.
// It returns error if the peer with that url already exists.
func (c *Channel) AddPeer(peer fab.Peer) error {
	url := peer.URL()
	if c.peers[url] != nil {
		return errors.Errorf("peer with URL %s already exists", url)
	}
	c.peers[url] = peer
	return nil
}

// RemovePeer remove a peer endpoint from channel.
func (c *Channel) RemovePeer(peer fab.Peer) {
	url := peer.URL()
	if c.peers[url] != nil {
		delete(c.peers, url)
		logger.Debugf("Removed peer with URL %s", url)
	}
}

// Peers returns the peers of of the channel.
func (c *Channel) Peers() []fab.Peer {
	var peersArray []fab.Peer
	for _, v := range c.peers {
		peersArray = append(peersArray, v)
	}
	return peersArray
}

// AnchorPeers returns the anchor peers for this channel.
// Note: channel.Initialize() must be called first to retrieve anchor peers
func (c *Channel) AnchorPeers() []fab.OrgAnchorPeer {
	anchors := []fab.OrgAnchorPeer{}
	for _, anchor := range c.anchorPeers {
		anchors = append(anchors, *anchor)
	}

	return anchors
}

// SetPrimaryPeer sets the primary peer -- The peer to use for doing queries.
// Peer must be a peer on this channel's peer list.
// Default: When no primary peer has been set the first peer
// on the list will be used.
// It returns error when peer is not on the existing peer list
func (c *Channel) SetPrimaryPeer(peer fab.Peer) error {

	if !c.isValidPeer(peer) {
		return errors.New("primary peer must be on this channel peer list")
	}

	c.primaryPeer = c.peers[peer.URL()]
	return nil
}

// PrimaryPeer gets the primary peer -- the peer to use for doing queries.
// Default: When no primary peer has been set the first peer
// from map range will be used.
func (c *Channel) PrimaryPeer() fab.Peer {

	if c.primaryPeer != nil {
		return c.primaryPeer
	}

	// When no primary peer has been set default to the first peer
	// from map range - order is not guaranteed
	for _, peer := range c.peers {
		logger.Debugf("Primary peer was not set, using %s", peer.URL())
		return peer
	}

	return nil
}

// AddOrderer adds an orderer endpoint to a channel object, this is a local-only operation.
// A channel instance may choose to use a single orderer node, which will broadcast
// requests to the rest of the orderer network. Or if the application does not trust
// the orderer nodes, it can choose to use more than one by adding them to the channel instance.
// All APIs concerning the orderer will broadcast to all orderers simultaneously.
// orderer: An instance of the Orderer interface.
// Returns error if the orderer with that url already exists.
func (c *Channel) AddOrderer(orderer fab.Orderer) error {
	url := orderer.URL()
	if c.orderers[url] != nil {
		return errors.Errorf("orderer with URL %s already exists", url)
	}
	c.orderers[orderer.URL()] = orderer
	return nil
}

// RemoveOrderer removes orderer endpoint from a channel object, this is a local-only operation.
// orderer: An instance of the Orderer class.
func (c *Channel) RemoveOrderer(orderer fab.Orderer) {
	url := orderer.URL()
	if c.orderers[url] != nil {
		delete(c.orderers, url)
		logger.Debugf("Removed orderer with URL %s", url)
	}
}

// Orderers gets the orderers of a channel.
func (c *Channel) Orderers() []fab.Orderer {
	var orderersArray []fab.Orderer
	for _, v := range c.orderers {
		orderersArray = append(orderersArray, v)
	}
	return orderersArray
}

// SetMSPManager sets the MSP Manager for this channel.
// This utility method will not normally be used as the
// "Initialize()" method will read this channel's
// current configuration and reset the MSPManager with
// the MSP's found.
func (c *Channel) SetMSPManager(mspManager msp.MSPManager) {
	c.mspManager = mspManager
}

// MSPManager returns the MSP Manager for this channel
func (c *Channel) MSPManager() msp.MSPManager {
	return c.mspManager
}

// OrganizationUnits - to get identifier for the organization configured on the channel
func (c *Channel) OrganizationUnits() ([]string, error) {
	channelMSPManager := c.MSPManager()
	msps, err := channelMSPManager.GetMSPs()
	if err != nil {
		logger.Info("Cannot get channel manager")
		return nil, errors.WithMessage(err, "organization units were not set")
	}
	var orgIdentifiers []string
	for _, v := range msps {
		orgName, err := v.GetIdentifier()
		if err != nil {
			logger.Info("Organization does not have an identifier")
		}
		orgIdentifiers = append(orgIdentifiers, orgName)
	}
	return orgIdentifiers, nil
}

// Utility function to ensure that a peer exists on this channel.
// It returns true if peer exists on this channel
func (c *Channel) isValidPeer(peer fab.Peer) bool {
	return peer != nil && c.peers[peer.URL()] != nil
}

// TODO
// The following functions haven't been implemented.

// UpdateChannel calls the orderer(s) to update an existing channel. This allows the addition and
// deletion of Peer nodes to an existing channel, as well as the update of Peer
// certificate information upon certificate renewals.
// It returns whether or not the channel update process was successful.
func (c *Channel) UpdateChannel() bool {
	return false
}

// IsReadonly gets channel status to see if the underlying channel has been terminated,
// making it a read-only channel, where information (transactions and states)
// can be queried but no new transactions can be submitted.
// It returns read-only, true or not.
func (c *Channel) IsReadonly() bool {
	return false //to do
}

// IsInitialized ... TODO
func (c *Channel) IsInitialized() bool {
	return c.initialized
}

//addCertsToConfig adds cert bytes to config TLSCACertPool
func addCertsToConfig(config core.Config, pemCerts []byte) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		config.TLSCACertPool(cert)
	}
}

// getOrdererConfig returns ordererconfig for given ordererAddress (ports will be ignored)
func getOrdererConfig(config core.Config, ordererAddress string) (*core.OrdererConfig, error) {
	return config.OrdererConfig(resolveOrdererAddress(ordererAddress))
}

// resolveOrdererAddress resolves order address to remove port from address if present
func resolveOrdererAddress(ordererAddress string) string {
	s := strings.Split(ordererAddress, ":")
	if len(s) > 1 {
		return s[0]
	}
	return ordererAddress
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
// This query will be made to the primary peer.
func (c *Channel) QueryInfo() (*common.BlockchainInfo, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	resps, err := l.QueryInfo([]fab.ProposalProcessor{c.PrimaryPeer()})
	if err != nil {
		return nil, err
	}
	return resps[0].BCI, err
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to the primary peer.
// Returns the block.
func (c *Channel) QueryBlockByHash(blockHash []byte) (*common.Block, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	resps, err := l.QueryBlockByHash(blockHash, []fab.ProposalProcessor{c.PrimaryPeer()})
	if err != nil {
		return nil, err
	}
	return resps[0], err
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to the primary peer.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func (c *Channel) QueryBlock(blockNumber int) (*common.Block, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	resps, err := l.QueryBlock(blockNumber, []fab.ProposalProcessor{c.PrimaryPeer()})
	if err != nil {
		return nil, err
	}
	return resps[0], err
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to the primary peer.
// Returns the ProcessedTransaction information containing the transaction.
// TODO: add optional target
func (c *Channel) QueryTransaction(transactionID fab.TransactionID) (*pb.ProcessedTransaction, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	resps, err := l.QueryTransaction(transactionID, []fab.ProposalProcessor{c.PrimaryPeer()})
	if err != nil {
		return nil, err
	}
	return resps[0], err
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on this channel.
// This query will be made to the primary peer.
func (c *Channel) QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	resps, err := l.QueryInstantiatedChaincodes([]fab.ProposalProcessor{c.PrimaryPeer()})
	if err != nil {
		return nil, err
	}
	return resps[0], err

}

// QueryConfigBlock returns the current configuration block for the specified channel. If the
// peer doesn't belong to the channel, return error
func (c *Channel) QueryConfigBlock(targets []fab.ProposalProcessor, minResponses int) (*common.ConfigEnvelope, error) {
	l, err := NewLedger(c.clientContext, c.name)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	return l.QueryConfigBlock(targets, minResponses)
}

// QueryByChaincode sends a proposal to one or more endorsing peers that will be handled by the chaincode.
// This request will be presented to the chaincode 'invoke' and must understand
// from the arguments that this is a query request. The chaincode must also return
// results in the byte array format and the caller will have to be able to decode.
// these results.
func (c *Channel) QueryByChaincode(request fab.ChaincodeInvokeRequest) ([][]byte, error) {
	targets, err := c.chaincodeInvokeRequestAddDefaultPeers(request.Targets)
	if err != nil {
		return nil, err
	}

	resps, err := queryChaincode(c.clientContext, c.name, request, targets)
	return collectProposalResponses(resps), err
}

// QueryBySystemChaincode invokes a chaincode that isn't part of a channel.
//
// TODO: This function's name is confusing - call the normal QueryByChaincode for system chaincode on a channel.
func (c *Channel) QueryBySystemChaincode(request fab.ChaincodeInvokeRequest) ([][]byte, error) {
	targets, err := c.chaincodeInvokeRequestAddDefaultPeers(request.Targets)
	if err != nil {
		return nil, err
	}

	resps, err := queryChaincode(c.clientContext, fab.SystemChannel, request, targets)
	return collectProposalResponses(resps), err
}

// SendInstantiateProposal sends an instantiate proposal to one or more endorsing peers.
func (c *Channel) SendInstantiateProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	if chaincodeName == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodeName is required")
	}
	if chaincodePath == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodePath is required")
	}
	if chaincodeVersion == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodeVersion is required")
	}
	if chaincodePolicy == nil {
		return nil, fab.EmptyTransactionID, errors.New("chaincodePolicy is required")
	}
	if len(targets) == 0 {
		return nil, fab.EmptyTransactionID, errors.New("missing peer objects for chaincode proposal")
	}

	cp := ChaincodeDeployRequest{
		Name:       chaincodeName,
		Args:       args,
		Path:       chaincodePath,
		Version:    chaincodeVersion,
		Policy:     chaincodePolicy,
		CollConfig: collConfig,
	}

	txh, err := txn.NewHeader(c.clientContext, c.name)
	if err != nil {
		return nil, fab.EmptyTransactionID, errors.WithMessage(err, "create transaction ID failed")
	}

	tp, err := CreateChaincodeDeployProposal(txh, InstantiateChaincode, c.name, cp)
	if err != nil {
		return nil, fab.EmptyTransactionID, errors.WithMessage(err, "creation of chaincode proposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
	return tpr, tp.TxnID, err
}

// SendUpgradeProposal sends an upgrade proposal to one or more endorsing peers.
func (c *Channel) SendUpgradeProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	if chaincodeName == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodeName is required")
	}
	if chaincodePath == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodePath is required")
	}
	if chaincodeVersion == "" {
		return nil, fab.EmptyTransactionID, errors.New("chaincodeVersion is required")
	}
	if chaincodePolicy == nil {
		return nil, fab.EmptyTransactionID, errors.New("chaincodePolicy is required")
	}
	if len(targets) == 0 {
		return nil, fab.EmptyTransactionID, errors.New("missing peer objects for chaincode proposal")
	}

	cp := ChaincodeDeployRequest{
		Name:    chaincodeName,
		Args:    args,
		Path:    chaincodePath,
		Version: chaincodeVersion,
		Policy:  chaincodePolicy,
	}

	txh, err := txn.NewHeader(c.clientContext, c.name)
	if err != nil {
		return nil, fab.EmptyTransactionID, errors.WithMessage(err, "create transaction ID failed")
	}

	tp, err := CreateChaincodeDeployProposal(txh, UpgradeChaincode, c.name, cp)
	if err != nil {
		return nil, fab.EmptyTransactionID, errors.WithMessage(err, "creation of chaincode proposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
	return tpr, tp.TxnID, err
}

func validateChaincodeInvokeRequest(request fab.ChaincodeInvokeRequest) error {
	if request.ChaincodeID == "" {
		return errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return errors.New("Fcn is required")
	}
	return nil
}

func (c *Channel) chaincodeInvokeRequestAddDefaultPeers(targets []fab.ProposalProcessor) ([]fab.ProposalProcessor, error) {
	// Use default peers if targets are not specified.
	if targets == nil || len(targets) == 0 {
		if c.peers == nil || len(c.peers) == 0 {
			return nil, status.New(status.ClientStatus, status.NoPeersFound.ToInt32(),
				"targets were not specified and no peers have been configured", nil)
		}

		return peersToTxnProcessors(c.Peers()), nil
	}
	return targets, nil
}

// block retrieves the block at the given position
func (c *Channel) block(pos *ab.SeekPosition) (*common.Block, error) {

	th, err := txn.NewHeader(c.clientContext, c.name)
	if err != nil {
		return nil, errors.Wrap(err, "generating TX ID failed")
	}

	channelHeaderOpts := txn.ChannelHeaderOpts{
		TxnHeader:   th,
		TLSCertHash: ccomm.TLSCertHash(c.clientContext.Config()),
	}
	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "CreateChannelHeader failed")
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	signatureHeader, err := txn.CreateSignatureHeader(th)
	if err != nil {
		return nil, errors.Wrap(err, "CreateSignatureHeader failed")
	}

	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signature header failed")
	}

	seekHeader := &common.Header{
		ChannelHeader:   seekInfoHeaderBytes,
		SignatureHeader: signatureHeaderBytes,
	}

	seekInfo := &ab.SeekInfo{
		Start:    pos,
		Stop:     pos,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	seekInfoBytes, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	payload := common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	return txn.SendPayload(c.clientContext, &payload, c.Orderers())
}

// newNewestSeekPosition returns a SeekPosition that requests the newest block
func newNewestSeekPosition() *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}}}
}

// newSpecificSeekPosition returns a SeekPosition that requests the block at the given index
func newSpecificSeekPosition(index uint64) *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: index}}}
}

// ChannelConfig queries for the current config block for this channel.
// This transaction will be made to the orderer.
// @returns {ConfigEnvelope} Object containing the configuration items.
// @see /protos/orderer/ab.proto
// @see /protos/common/configtx.proto
func (c *Channel) ChannelConfig() (*common.ConfigEnvelope, error) {
	logger.Debugf("channelConfig - start for channel %s", c.name)

	// Get the newest block
	block, err := c.block(newNewestSeekPosition())
	if err != nil {
		return nil, err
	}
	logger.Debugf("channelConfig - Retrieved newest block number: %d\n", block.Header.Number)

	// Get the index of the last config block
	lastConfig, err := getLastConfigFromBlock(block)
	if err != nil {
		return nil, errors.Wrap(err, "GetLastConfigFromBlock failed")
	}
	logger.Debugf("channelConfig - Last config index: %d\n", lastConfig.Index)

	// Get the last config block
	block, err = c.block(newSpecificSeekPosition(lastConfig.Index))

	if err != nil {
		return nil, errors.WithMessage(err, "retrieve block failed")
	}
	logger.Debugf("channelConfig - Last config block number %d, Number of tx: %d", block.Header.Number, len(block.Data.Data))

	if len(block.Data.Data) != 1 {
		return nil, errors.New("apiconfig block must contain one transaction")
	}

	return createConfigEnvelope(block.Data.Data[0])

}

func loadMSPs(mspConfigs []*mb.MSPConfig, cs core.CryptoSuite) ([]msp.MSP, error) {
	logger.Debugf("loadMSPs - start number of msps=%d", len(mspConfigs))

	msps := []msp.MSP{}
	for _, config := range mspConfigs {
		mspType := msp.ProviderType(config.Type)
		if mspType != msp.FABRIC {
			return nil, errors.Errorf("MSP type not supported: %v", mspType)
		}
		if len(config.Config) == 0 {
			return nil, errors.Errorf("MSP configuration missing the payload in the 'Config' property")
		}

		fabricConfig := &mb.FabricMSPConfig{}
		err := proto.Unmarshal(config.Config, fabricConfig)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal FabricMSPConfig from config failed")
		}

		if fabricConfig.Name == "" {
			return nil, errors.New("MSP Configuration missing name")
		}

		// with this method we are only dealing with verifying MSPs, not local MSPs. Local MSPs are instantiated
		// from user enrollment materials (see User class). For verifying MSPs the root certificates are always
		// required
		if len(fabricConfig.RootCerts) == 0 {
			return nil, errors.New("MSP Configuration missing root certificates required for validating signing certificates")
		}

		// get the application org names
		var orgs []string
		orgUnits := fabricConfig.OrganizationalUnitIdentifiers
		for _, orgUnit := range orgUnits {
			logger.Debugf("loadMSPs - found org of :: %s", orgUnit.OrganizationalUnitIdentifier)
			orgs = append(orgs, orgUnit.OrganizationalUnitIdentifier)
		}

		// TODO: Do something with orgs
		// TODO: Configure MSP version (rather than MSP 1.0)
		newMSP, err := msp.NewBccspMsp(msp.MSPv1_0, cs)
		if err != nil {
			return nil, errors.Wrap(err, "instantiate MSP failed")
		}

		if err := newMSP.Setup(config); err != nil {
			return nil, errors.Wrap(err, "configure MSP failed")
		}

		mspID, _ := newMSP.GetIdentifier()
		logger.Debugf("loadMSPs - adding msp=%s", mspID)

		msps = append(msps, newMSP)
	}

	logger.Debugf("loadMSPs - loaded %d MSPs", len(msps))
	return msps, nil
}

// getLastConfigFromBlock returns the LastConfig data from the given block
func getLastConfigFromBlock(block *common.Block) (*common.LastConfig, error) {
	if block.Metadata == nil {
		return nil, errors.New("block metadata is nil")
	}
	metadata := &common.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[common.BlockMetadataIndex_LAST_CONFIG], metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal block metadata failed")
	}

	lastConfig := &common.LastConfig{}
	err = proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal last config from metadata failed")
	}

	return lastConfig, err
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

// CreateChaincodeDeployProposal creates an instantiate or upgrade chaincode proposal.
func CreateChaincodeDeployProposal(txh fab.TransactionHeader, deploy ChaincodeProposalType, channelID string, chaincode ChaincodeDeployRequest) (*fab.TransactionProposal, error) {

	// Generate arguments for deploy (channel, marshaled CCDS, marshaled chaincode policy, marshaled collection policy)
	args := [][]byte{}
	args = append(args, []byte(channelID))

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincode.Name, Path: chaincode.Path, Version: chaincode.Version},
		Input: &pb.ChaincodeInput{Args: chaincode.Args}}}
	ccdsBytes, err := protos_utils.Marshal(ccds)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	chaincodePolicyBytes, err := protos_utils.Marshal(chaincode.Policy)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode policy failed")
	}
	args = append(args, chaincodePolicyBytes)

	args = append(args, []byte(escc))
	args = append(args, []byte(vscc))

	if chaincode.CollConfig != nil {
		var err error
		collConfigBytes, err := proto.Marshal(&common.CollectionConfigPackage{Config: chaincode.CollConfig})
		if err != nil {
			return nil, errors.WithMessage(err, "marshal of collection policy failed")
		}
		args = append(args, collConfigBytes)
	}

	// Fcn is deploy or upgrade
	fcn := ""
	switch deploy {
	case InstantiateChaincode:
		fcn = lsccDeploy
	case UpgradeChaincode:
		fcn = lsccUpgrade
	default:
		return nil, errors.WithMessage(err, "chaincode deployment type unknown")
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         fcn,
		Args:        args,
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}
