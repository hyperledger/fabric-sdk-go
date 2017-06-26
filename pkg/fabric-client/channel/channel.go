/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"time"

	"github.com/golang/protobuf/proto"
	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"

	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"

	proto_ts "github.com/golang/protobuf/ptypes/timestamp"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"

	fabric_config "github.com/hyperledger/fabric/common/config"
	mb "github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

type channel struct {
	name            string // Name of the channel is only meaningful to the client
	securityEnabled bool   // Security enabled flag
	peers           map[string]api.Peer
	tcertBatchSize  int // The number of tcerts to get in each batch
	orderers        map[string]api.Orderer
	clientContext   api.FabricClient
	primaryPeer     api.Peer
	mspManager      msp.MSPManager
	anchorPeers     []*api.OrgAnchorPeer
}

// configItems contains the configuration values retrieved from the Orderer Service
type configItems struct {
	msps        []*mb.MSPConfig
	anchorPeers []*api.OrgAnchorPeer
	orderers    []string
	versions    *versions
}

// versions ...
type versions struct {
	ReadSet  *common.ConfigGroup
	WriteSet *common.ConfigGroup
	Channel  *common.ConfigGroup
}

// NewChannel ...
/**
* @param {string} name to identify different channel instances. The naming of channel instances
* is enforced by the ordering service and must be unique within the blockchain network
* @param {FabricClient} clientContext An instance of {@link FabricClient} that provides operational context
* such as submitting User etc.
 */
func NewChannel(name string, client api.FabricClient) (api.Channel, error) {
	if name == "" {
		return nil, fmt.Errorf("failed to create Channel. Missing required 'name' parameter")
	}
	if client == nil {
		return nil, fmt.Errorf("failed to create Channel. Missing required 'clientContext' parameter")
	}
	p := make(map[string]api.Peer)
	o := make(map[string]api.Orderer)
	c := &channel{name: name, securityEnabled: client.GetConfig().IsSecurityEnabled(), peers: p,
		tcertBatchSize: client.GetConfig().TcertBatchSize(), orderers: o, clientContext: client, mspManager: msp.NewMSPManager()}
	logger.Infof("Constructed channel instance: %v", c)

	return c, nil
}

func (c *channel) QueryExtensionInterface() api.ChannelExtension {
	return c
}

// GetClientContext returns the Client that was passed in to NewChannel
func (c *channel) GetClientContext() api.FabricClient {
	return c.clientContext
}

// GetProposalBytes returns the serialized transaction.
func (c *channel) GetProposalBytes(tp *api.TransactionProposal) ([]byte, error) {
	return proto.Marshal(tp.SignedProposal)
}

// GetName ...
/**
* Get the channel name.
* @returns {string} The name of the channel.
 */
func (c *channel) GetName() string {
	return c.name
}

// IsSecurityEnabled ...
/**
* Determine if security is enabled.
 */
func (c *channel) IsSecurityEnabled() bool {
	return c.securityEnabled
}

// GetTCertBatchSize ...
/**
* Get the tcert batch size.
 */
func (c *channel) GetTCertBatchSize() int {
	return c.tcertBatchSize
}

// SetTCertBatchSize ...
/**
* Set the tcert batch size.
 */
func (c *channel) SetTCertBatchSize(batchSize int) {
	c.tcertBatchSize = batchSize
}

// AddPeer ...
/**
* Add peer endpoint to channel.
* @param {Peer} peer An instance of the Peer class that has been initialized with URL,
* TLC certificate, and enrollment certificate.
* @throws {Error} if the peer with that url already exists.
 */
func (c *channel) AddPeer(peer api.Peer) error {
	url := peer.URL()
	if c.peers[url] != nil {
		return fmt.Errorf("Peer with URL %s already exists", url)
	}
	c.peers[url] = peer
	return nil
}

// RemovePeer ...
/**
* Remove peer endpoint from channel.
* @param {Peer} peer An instance of the Peer.
 */
func (c *channel) RemovePeer(peer api.Peer) {
	url := peer.URL()
	if c.peers[url] != nil {
		delete(c.peers, url)
		logger.Debugf("Removed peer with URL %s", url)
	}
}

// GetPeers ...
/**
* Get peers of a channel from local information.
* @returns {[]Peer} The peer list on the channel.
 */
func (c *channel) GetPeers() []api.Peer {
	var peersArray []api.Peer
	for _, v := range c.peers {
		peersArray = append(peersArray, v)
	}
	return peersArray
}

// GetAnchorPeers returns the anchor peers for this channel.
// Note: channel.Initialize() must be called first to retrieve anchor peers
func (c *channel) GetAnchorPeers() []api.OrgAnchorPeer {
	anchors := []api.OrgAnchorPeer{}
	for _, anchor := range c.anchorPeers {
		anchors = append(anchors, *anchor)
	}

	return anchors
}

/**
* Utility function to get target peers (target peer is valid only if it belongs to channel's peer list).
* If targets is empty return channel's peer list
* @returns {[]Peer} The target peer list
* @returns {error} if target peer is not in channel's peer list
 */
func (c *channel) getTargetPeers(targets []api.Peer) ([]api.Peer, error) {

	if targets == nil || len(targets) == 0 {
		return c.GetPeers(), nil
	}

	var targetPeers []api.Peer
	for _, target := range targets {
		if !c.isValidPeer(target) {
			return nil, fmt.Errorf("The target peer must be on this channel peer list")
		}
		targetPeers = append(targetPeers, c.peers[target.URL()])
	}

	return targetPeers, nil
}

/**
* Utility function to ensure that a peer exists on this channel
* @returns {bool} true if peer exists on this channel
 */
func (c *channel) isValidPeer(peer api.Peer) bool {
	return peer != nil && c.peers[peer.URL()] != nil
}

// SetPrimaryPeer ...
/**
* Set the primary peer
* The peer to use for doing queries.
* Peer must be a peer on this channel's peer list.
* Default: When no primary peer has been set the first peer
* on the list will be used.
* @param {Peer} peer An instance of the Peer class.
* @returns error when peer is not on the existing peer list
 */
func (c *channel) SetPrimaryPeer(peer api.Peer) error {

	if !c.isValidPeer(peer) {
		return fmt.Errorf("The primary peer must be on this channel peer list")
	}

	c.primaryPeer = c.peers[peer.URL()]
	return nil
}

// GetPrimaryPeer ...
/**
* Get the primary peer
* The peer to use for doing queries.
* Default: When no primary peer has been set the first peer
* from map range will be used.
* @returns {Peer} peer An instance of the Peer class.
 */
func (c *channel) GetPrimaryPeer() api.Peer {

	if c.primaryPeer != nil {
		return c.primaryPeer
	}

	// When no primary peer has been set default to the first peer
	// from map range - order is not guaranteed
	for _, peer := range c.peers {
		logger.Infof("Primary peer was not set, using %s", peer.Name())
		return peer
	}

	return nil
}

// AddOrderer ...
/**
* Add orderer endpoint to a channel object, this is a local-only operation.
* A channel instance may choose to use a single orderer node, which will broadcast
* requests to the rest of the orderer network. Or if the application does not trust
* the orderer nodes, it can choose to use more than one by adding them to the channel instance.
* All APIs concerning the orderer will broadcast to all orderers simultaneously.
* @param {Orderer} orderer An instance of the Orderer class.
* @throws {Error} if the orderer with that url already exists.
 */
func (c *channel) AddOrderer(orderer api.Orderer) error {
	url := orderer.GetURL()
	if c.orderers[url] != nil {
		return fmt.Errorf("Orderer with URL %s already exists", url)
	}
	c.orderers[orderer.GetURL()] = orderer
	return nil
}

// RemoveOrderer ...
/**
* Remove orderer endpoint from a channel object, this is a local-only operation.
* @param {Orderer} orderer An instance of the Orderer class.
 */
func (c *channel) RemoveOrderer(orderer api.Orderer) {
	url := orderer.GetURL()
	if c.orderers[url] != nil {
		delete(c.orderers, url)
		logger.Debugf("Removed orderer with URL %s", url)
	}
}

// GetOrderers ...
/**
* Get orderers of a channel.
 */
func (c *channel) GetOrderers() []api.Orderer {
	var orderersArray []api.Orderer
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
func (c *channel) SetMSPManager(mspManager msp.MSPManager) {
	c.mspManager = mspManager
}

// GetMSPManager returns the MSP Manager for this channel
func (c *channel) GetMSPManager() msp.MSPManager {
	return c.mspManager
}

// GetOrganizationUnits - to get identifier for the organization configured on the channel
func (c *channel) GetOrganizationUnits() ([]string, error) {
	channelMSPManager := c.GetMSPManager()
	msps, err := channelMSPManager.GetMSPs()
	if err != nil {
		logger.Info("Cannot get channel manager")
		return nil, fmt.Errorf("Organization uits were not set: %v", err)
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

// GetGenesisBlock ...
/**
* Will get the genesis block from the defined orderer that may be
* used in a join request
* @param {Object} request - An object containing the following fields:
*                            <br>`txId` : required - String of the transaction id
*                            <br>`nonce` : required - Integer of the once time number
*
* @returns A Genesis block
* @see /protos/peer/proposal_response.proto
 */
func (c *channel) GetGenesisBlock(request *api.GenesisBlockRequest) (*common.Block, error) {
	logger.Debug("GetGenesisBlock - start")

	// verify that we have an orderer configured
	if len(c.GetOrderers()) == 0 {
		return nil, fmt.Errorf("GetGenesisBlock - error: Missing orderer assigned to this channel for the getGenesisBlock request")
	}
	// verify that we have transaction id
	if request.TxID == "" {
		return nil, fmt.Errorf("GetGenesisBlock - error: Missing txId input parameter with the required transaction identifier")
	}
	// verify that we have the nonce
	if request.Nonce == nil {
		return nil, fmt.Errorf("GetGenesisBlock - error: Missing nonce input parameter with the required single use number")
	}

	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("Error getting creator: %v", err)
	}

	// now build the seek info , will be used once the channel is created
	// to get the genesis block back
	seekStart := fc.NewSpecificSeekPosition(0)
	seekStop := fc.NewSpecificSeekPosition(0)
	seekInfo := &ab.SeekInfo{
		Start:    seekStart,
		Stop:     seekStop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}
	protos_utils.MakeChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, 1, c.GetName(), 0)
	seekInfoHeader, err := BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, c.GetName(), request.TxID, 0, "", time.Now())
	if err != nil {
		return nil, fmt.Errorf("Error building channel header: %v", err)
	}
	seekHeader, err := fc.BuildHeader(creator, seekInfoHeader, request.Nonce)
	if err != nil {
		return nil, fmt.Errorf("Error building header: %v", err)
	}
	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   fc.MarshalOrPanic(seekInfo),
	}
	seekPayloadBytes := fc.MarshalOrPanic(seekPayload)

	signedEnvelope, err := c.SignPayload(seekPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("Error signing payload: %v", err)
	}

	block, err := c.SendEnvelope(signedEnvelope)
	if err != nil {
		return nil, fmt.Errorf("Error sending envelope: %v", err)
	}
	return block, nil
}

// JoinChannel ...
/**
* Sends a join channel proposal to one or more endorsing peers
* Will get the genesis block from the defined orderer to be used
* in the proposal.
* @param {Object} request - An object containing the following fields:
*   <br>`targets` : required - An array of `Peer` objects that will join
*                   this channel
*   <br>`block` : the genesis block of the channel
*                 see getGenesisBlock() method
*   <br>`txId` : required - String of the transaction id
*   <br>`nonce` : required - Integer of the once time number
* @returns {Promise} A Promise for a `ProposalResponse`
* @see /protos/peer/proposal_response.proto
 */
func (c *channel) JoinChannel(request *api.JoinChannelRequest) error {
	logger.Debug("joinChannel - start")

	// verify that we have targets (Peers) to join this channel
	// defined by the caller
	if request == nil {
		return fmt.Errorf("JoinChannel - error: Missing all required input request parameters")
	}

	// verify that a Peer(s) has been selected to join this channel
	if request.Targets == nil {
		return fmt.Errorf("JoinChannel - error: Missing targets input parameter with the peer objects for the join channel proposal")
	}

	// verify that we have transaction id
	if request.TxID == "" {
		return fmt.Errorf("JoinChannel - error: Missing txId input parameter with the required transaction identifier")
	}

	// verify that we have the nonce
	if request.Nonce == nil {
		return fmt.Errorf("JoinChannel - error: Missing nonce input parameter with the required single use number")
	}

	if request.GenesisBlock == nil {
		return fmt.Errorf("JoinChannel - error: Missing block input parameter with the required genesis block")
	}

	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return fmt.Errorf("Error getting creator ID: %v", err)
	}

	genesisBlockBytes, err := proto.Marshal(request.GenesisBlock)
	if err != nil {
		return fmt.Errorf("Error marshalling genesis block: %v", err)
	}

	// Create join channel transaction proposal for target peers
	joinCommand := "JoinChain"
	var args [][]byte
	args = append(args, []byte(joinCommand))
	args = append(args, genesisBlockBytes)
	ccSpec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_GOLANG,
		ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cciSpec := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: ccSpec,
	}

	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(request.TxID, common.HeaderType_ENDORSER_TRANSACTION, "", cciSpec, request.Nonce, creator, nil)
	if err != nil {
		return fmt.Errorf("Error building proposal: %v", err)
	}
	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return fmt.Errorf("Error signing proposal: %v", err)
	}
	transactionProposal := &api.TransactionProposal{
		TransactionID:  txID,
		SignedProposal: signedProposal,
		Proposal:       proposal,
	}

	// Send join proposal
	proposalResponses, err := c.SendTransactionProposal(transactionProposal, 0, request.Targets)
	if err != nil {
		return fmt.Errorf("Error sending join transaction proposal: %s", err)
	}
	// Check responses from target peers for success/failure and join all errors
	var joinError string
	for _, response := range proposalResponses {
		if response.Err != nil {
			joinError = joinError +
				fmt.Sprintf("Join channel proposal response error: %s \n",
					response.Err.Error())
		} else if response.Status != http.StatusOK {
			joinError = joinError +
				fmt.Sprintf("Join channel proposal HTTP response error: %s \n",
					response.Err.Error())
		}
	}

	if joinError != "" {
		return fmt.Errorf(joinError)
	}

	return nil
}

/**
* Queries for the current config block for this channel.
* This transaction will be made to the orderer.
* @returns {ConfigEnvelope} Object containing the configuration items.
* @see /protos/orderer/ab.proto
* @see /protos/common/configtx.proto
 */
func (c *channel) getChannelConfig() (*common.ConfigEnvelope, error) {
	logger.Debugf("getChannelConfig - start for channel %s", c.name)

	// Get the newest block
	block, err := c.getBlock(fc.NewNewestSeekPosition())
	if err != nil {
		return nil, err
	}
	logger.Debugf("GetChannelConfig - Retrieved newest block number: %d\n", block.Header.Number)

	// Get the index of the last config block
	lastConfig, err := fc.GetLastConfigFromBlock(block)
	if err != nil {
		return nil, fmt.Errorf("Unable to get last config from block: %v", err)
	}
	logger.Debugf("GetChannelConfig - Last config index: %d\n", lastConfig.Index)

	// Get the last config block
	//block, err = c.getBlock(NewSpecificSeekPosition(lastConfig.Index))
	block, err = c.getBlock(fc.NewSpecificSeekPosition(0)) //FIXME: temporary hack to workaround https://jira.hyperledger.org/browse/FAB-3493
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve block at index %d: %v", lastConfig.Index, err)
	}
	logger.Debugf("GetChannelConfig - Last config block number %d, Number of tx: %d", block.Header.Number, len(block.Data.Data))

	if len(block.Data.Data) != 1 {
		return nil, fmt.Errorf("Config block must only contain one transaction but contains %d", len(block.Data.Data))
	}

	envelope := &common.Envelope{}
	if err = proto.Unmarshal(block.Data.Data[0], envelope); err != nil {
		return nil, fmt.Errorf("Error extracting envelope from config block: %v", err)
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, fmt.Errorf("Error extracting payload from envelope: %s", err)
	}
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, fmt.Errorf("Error extracting payload from envelope: %s", err)
	}
	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, fmt.Errorf("Block must be of type 'CONFIG'")
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal config envelope: %v", err)
	}
	return configEnvelope, nil
}

// LoadConfigUpdateEnvelope ...
/*
* Utility method to load this channel with configuration information
* from an Envelope that contains a Configuration
* @param {byte[]} the envelope with the configuration update items
* @see /protos/common/configtx.proto
 */
func (c *channel) LoadConfigUpdateEnvelope(data []byte) error {
	logger.Debugf("loadConfigUpdateEnvelope - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(data, envelope)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal envelope: %v", err)
	}

	payload, err := protos_utils.ExtractPayload(envelope)
	if err != nil {
		return fmt.Errorf("Unable to extract payload from config update envelope: %v", err)
	}

	channelHeader, err := protos_utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return fmt.Errorf("Unable to extract channel header from config update payload: %v", err)
	}

	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG_UPDATE {
		return fmt.Errorf("Block must be of type 'CONFIG_UPDATE'")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	if err := proto.Unmarshal(payload.Data, configUpdateEnvelope); err != nil {
		return fmt.Errorf("Unable to unmarshal config update envelope: %v", err)
	}

	_, err = c.loadConfigUpdate(configUpdateEnvelope.ConfigUpdate)
	return err
}

func (c *channel) loadConfigUpdate(configUpdateBytes []byte) (*configItems, error) {

	configUpdate := &common.ConfigUpdate{}
	if err := proto.Unmarshal(configUpdateBytes, configUpdate); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal config update: %v", err)
	}
	logger.Debugf("loadConfigUpdate - channel ::" + configUpdate.ChannelId)

	readSet := configUpdate.ReadSet
	writeSet := configUpdate.WriteSet

	versions := &versions{
		ReadSet:  readSet,
		WriteSet: writeSet,
	}

	configItems := &configItems{
		msps:        []*mb.MSPConfig{},
		anchorPeers: []*api.OrgAnchorPeer{},
		orderers:    []string{},
		versions:    versions,
	}

	err := loadConfigGroup(configItems, configItems.versions.ReadSet, readSet, "read_set", "", false)
	if err != nil {
		return nil, err
	}
	// do the write_set second so they update anything in the read set
	err = loadConfigGroup(configItems, configItems.versions.WriteSet, writeSet, "write_set", "", false)
	if err != nil {
		return nil, err
	}
	err = c.initializeFromConfig(configItems)
	if err != nil {
		return nil, fmt.Errorf("channel initialization errort: %v", err)
	}

	//TODO should we create orderers and endorsing peers
	return configItems, nil
}

func (c *channel) loadConfigEnvelope(configEnvelope *common.ConfigEnvelope) (*configItems, error) {

	group := configEnvelope.Config.ChannelGroup

	versions := &versions{
		Channel: &common.ConfigGroup{},
	}

	configItems := &configItems{
		msps:        []*mb.MSPConfig{},
		anchorPeers: []*api.OrgAnchorPeer{},
		orderers:    []string{},
		versions:    versions,
	}

	err := loadConfigGroup(configItems, configItems.versions.Channel, group, "base", "", true)
	if err != nil {
		return nil, fmt.Errorf("Unable to load config items from channel group: %v", err)
	}

	err = c.initializeFromConfig(configItems)

	logger.Debugf("channel config: %v", configItems)

	return configItems, err
}

// UpdateChannel ...
/**
* Calls the orderer(s) to update an existing channel. This allows the addition and
* deletion of Peer nodes to an existing channel, as well as the update of Peer
* certificate information upon certificate renewals.
* @returns {bool} Whether the channel update process was successful.
 */
func (c *channel) UpdateChannel() bool {
	return false
}

// IsReadonly ...
/**
* Get channel status to see if the underlying channel has been terminated,
* making it a read-only channel, where information (transactions and states)
* can be queried but no new transactions can be submitted.
* @returns {bool} Is read-only, true or not.
 */
func (c *channel) IsReadonly() bool {
	return false //to do
}

// QueryInfo ...
/**
* Queries for various useful information on the state of the channel
* (height, known peers).
* This query will be made to the primary peer.
* @returns {object} With height, currently the only useful info.
 */
func (c *channel) QueryInfo() (*common.BlockchainInfo, error) {
	logger.Debug("queryInfo - start")

	// prepare arguments to call qscc GetChainInfo function
	var args []string
	args = append(args, "GetChainInfo")
	args = append(args, c.GetName())

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetChainInfo return error: %v", err)
	}

	bci := &common.BlockchainInfo{}
	err = proto.Unmarshal(payload, bci)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal BlockchainInfo return error: %v", err)
	}

	return bci, nil
}

// QueryBlockByHash ...
/**
* Queries the ledger for Block by block hash.
* This query will be made to the primary peer.
* @param {byte[]} block hash of the Block.
* @returns {object} Object containing the block.
 */
func (c *channel) QueryBlockByHash(blockHash []byte) (*common.Block, error) {

	if blockHash == nil {
		return nil, fmt.Errorf("Blockhash bytes are required")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args []string
	args = append(args, "GetBlockByHash")
	args = append(args, c.GetName())
	args = append(args, string(blockHash[:len(blockHash)]))

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByHash return error: %v", err)
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal Block return error: %v", err)
	}

	return block, nil
}

// QueryBlock ...
/**
* Queries the ledger for Block by block number.
* This query will be made to the primary peer.
* @param {int} blockNumber The number which is the ID of the Block.
* @returns {object} Object containing the block.
 */
func (c *channel) QueryBlock(blockNumber int) (*common.Block, error) {

	if blockNumber < 0 {
		return nil, fmt.Errorf("Block number must be positive integer")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args []string
	args = append(args, "GetBlockByNumber")
	args = append(args, c.GetName())
	args = append(args, strconv.Itoa(blockNumber))

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByNumber return error: %v", err)
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal Block return error: %v", err)
	}

	return block, nil
}

// Initialize initializes the channel
/**
* Retrieve the configuration from the primary orderer and initializes this channel
* with those values. Optionally a configuration may be passed in to initialize this channel
* without making the call to the orderer.
* @param {byte[]} config_update- Optional - A serialized form of the protobuf configuration update
 */
func (c *channel) Initialize(configUpdate []byte) error {

	if len(configUpdate) > 0 {
		var err error
		if _, err = c.loadConfigUpdate(configUpdate); err != nil {
			return fmt.Errorf("Unable to load config update envelope: %v", err)
		}
		return nil
	}

	configEnvelope, err := c.getChannelConfig()
	if err != nil {
		return fmt.Errorf("Unable to retrieve channel configuration from orderer service: %v", err)
	}

	_, err = c.loadConfigEnvelope(configEnvelope)
	if err != nil {
		return fmt.Errorf("Unable to load config envelope: %v", err)
	}
	return nil
}

// QueryTransaction ...
/**
* Queries the ledger for Transaction by number.
* This query will be made to the primary peer.
* @param {int} transactionID
* @returns {object} ProcessedTransaction information containing the transaction.
 */
func (c *channel) QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error) {

	// prepare arguments to call qscc GetTransactionByID function
	var args []string
	args = append(args, "GetTransactionByID")
	args = append(args, c.GetName())
	args = append(args, transactionID)

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByNumber return error: %v", err)
	}

	transaction := new(pb.ProcessedTransaction)
	err = proto.Unmarshal(payload, transaction)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ProcessedTransaction return error: %v", err)
	}

	return transaction, nil
}

//QueryInstantiatedChaincodes
/**
* Queries the instantiated chaincodes on this channel.
* This query will be made to the primary peer.
* @returns {object} ChaincodeQueryResponse proto
 */
func (c *channel) QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error) {

	payload, err := c.queryByChaincodeByTarget("lscc", []string{"getchaincodes"}, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke lscc getchaincodes return error: %v", err)
	}

	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ChaincodeQueryResponse return error: %v", err)
	}

	return response, nil
}

/**
* Generic helper for query functionality for chain
* This query will be made to one target peer and will return one result only.
* @parame {string} chaincode name
* @param {[]string} invoke arguments
* @param {Peer} target peer
* @returns {[]byte} payload
 */
func (c *channel) queryByChaincodeByTarget(chaincodeName string, args []string, target api.Peer) ([]byte, error) {

	queryResponses, err := c.QueryByChaincode(chaincodeName, args, []api.Peer{target})
	if err != nil {
		return nil, fmt.Errorf("QueryByChaincode return error: %v", err)
	}

	// we are only querying one peer hence one result
	if len(queryResponses) != 1 {
		return nil, fmt.Errorf("queryByChaincodeByTarget should have one result only - result number: %d", len(queryResponses))
	}

	return queryResponses[0], nil

}

//QueryByChaincode ..
/**
* Sends a proposal to one or more endorsing peers that will be handled by the chaincode.
* This request will be presented to the chaincode 'invoke' and must understand
* from the arguments that this is a query request. The chaincode must also return
* results in the byte array format and the caller will have to be able to decode
* these results
* @parame {string} chaincode name
* @param {[]string} invoke arguments
* @param {[]Peer} target peers
* @param {FabricClient} fabric client
* @returns {[][]byte} an array of payloads
 */
func QueryByChaincode(chaincodeName string, args []string, targets []api.Peer, clientContext api.FabricClient) ([][]byte, error) {
	if chaincodeName == "" {
		return nil, fmt.Errorf("Missing chaincode name")
	}

	if args == nil || len(args) < 1 {
		return nil, fmt.Errorf("Missing invoke arguments")
	}

	if targets == nil || len(targets) < 1 {
		return nil, fmt.Errorf("Missing target peers")
	}

	logger.Debugf("Calling %s function %v on targets: %s\n", chaincodeName, args[0], targets)

	signedProposal, err := CreateTransactionProposal(chaincodeName, "", args, true, nil, clientContext)
	if err != nil {
		return nil, fmt.Errorf("CreateTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	var responses [][]byte
	errMsg := ""
	for _, response := range transactionProposalResponses {
		if response.Err != nil {
			errMsg = errMsg + response.Err.Error() + "\n"
		} else {
			responses = append(responses, response.ProposalResponse.GetResponse().Payload)
		}
	}

	if len(errMsg) > 0 {
		return responses, fmt.Errorf(errMsg)
	}

	return responses, nil
}

func (c *channel) QueryByChaincode(chaincodeName string, args []string, targets []api.Peer) ([][]byte, error) {
	return QueryByChaincode(chaincodeName, args, targets, c.clientContext)
}

// CreateTransactionProposal ...
/**
* Create  a proposal for transaction. This involves assembling the proposal
* with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
* ECert to sign.
 */
func (c *channel) CreateTransactionProposal(chaincodeName string, channelID string,
	args []string, sign bool, transientData map[string][]byte) (*api.TransactionProposal, error) {
	return CreateTransactionProposal(chaincodeName, channelID, args, sign, transientData, c.clientContext)
}

//CreateTransactionProposal  ...
func CreateTransactionProposal(chaincodeName string, channelID string,
	args []string, sign bool, transientData map[string][]byte, clientContext api.FabricClient) (*api.TransactionProposal, error) {

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	creator, err := clientContext.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("Error getting creator: %v", err)
	}

	// create a proposal from a ChaincodeInvocationSpec
	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTransient(common.HeaderType_ENDORSER_TRANSACTION, channelID, ccis, creator, transientData)
	if err != nil {
		return nil, fmt.Errorf("Could not create chaincode proposal, err %s", err)
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling proposal: %v", err)
	}

	user, err := clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("Error loading user from store: %s", err)
	}

	signature, err := fc.SignObjectWithKey(proposalBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil, clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}
	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	return &api.TransactionProposal{
		TransactionID:  txID,
		SignedProposal: signedProposal,
		Proposal:       proposal,
	}, nil

}

// SendTransactionProposal ...
// Send  the created proposal to peer for endorsement.
func (c *channel) SendTransactionProposal(proposal *api.TransactionProposal, retry int, targets []api.Peer) ([]*api.TransactionProposalResponse, error) {
	if c.peers == nil || len(c.peers) == 0 {
		return nil, fmt.Errorf("peers is nil")
	}
	if proposal == nil || proposal.SignedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	targetPeers, err := c.getTargetPeers(targets)
	if err != nil {
		return nil, fmt.Errorf("GetTargetPeers return error: %s", err)
	}
	if len(targetPeers) < 1 {
		return nil, fmt.Errorf("Missing peer objects for sending transaction proposal")
	}

	return SendTransactionProposal(proposal, retry, targetPeers)

}

//SendTransactionProposal ...
func SendTransactionProposal(proposal *api.TransactionProposal, retry int, targetPeers []api.Peer) ([]*api.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	if len(targetPeers) < 1 {
		return nil, fmt.Errorf("Missing peer objects for sending transaction proposal")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*api.TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range targetPeers {
		wg.Add(1)
		go func(peer api.Peer) {
			defer wg.Done()
			var err error
			var proposalResponse *api.TransactionProposalResponse
			logger.Debugf("Send ProposalRequest to peer :%s", peer.URL())
			if proposalResponse, err = peer.SendProposal(proposal); err != nil {
				logger.Debugf("Receive Error Response :%v", proposalResponse)
				proposalResponse = &api.TransactionProposalResponse{
					Endorser: peer.URL(),
					Err:      fmt.Errorf("Error calling endorser '%s':  %s", peer.URL(), err),
					Proposal: proposal,
				}
			} else {
				logger.Debugf("Receive Proposal ChaincodeActionResponse :%v\n", proposalResponse)
			}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, proposalResponse)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, nil
}

// CreateTransaction ...
/**
* Create a transaction with proposal response, following the endorsement policy.
 */
func (c *channel) CreateTransaction(resps []*api.TransactionProposalResponse) (*api.Transaction, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("At least one proposal response is necessary")
	}

	proposal := resps[0].Proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Proposal.Payload)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal payload")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := protos_utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	// This code is commented out because the ProposalResponsePayload Extension ChaincodeAction Results
	// return from endorsements is different so the compare will fail

	//            var a1 []byte
	//            for n, r := range resps {
	//                            if n == 0 {
	//                                            a1 = r.Payload
	//                                            if r.Response.Status != 200 {
	//                                                            return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
	//                                            }
	//                                            continue
	//                            }

	//                            if bytes.Compare(a1, r.Payload) != 0 {
	//                                            return nil, fmt.Errorf("ProposalResponsePayloads do not match")
	//                            }
	//            }

	for _, r := range resps {
		if r.ProposalResponse.Response.Status != 200 {
			return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.ProposalResponse.Endorsement
	}
	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].ProposalResponse.Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protos_utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protos_utils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pb.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pb.TransactionAction, 1)
	taas[0] = taa

	return &api.Transaction{
		Transaction: &pb.Transaction{Actions: taas},
		Proposal:    proposal,
	}, nil
}

// SendTransaction ...
/**
* Send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
* This call is asynchronous and the successful transaction commit is notified via a BLOCK or CHAINCODE event. This method must provide a mechanism for applications to attach event listeners to handle “transaction submitted”, “transaction complete” and “error” events.
* Note that under the cover there are two different kinds of communications with the fabric backend that trigger different events to
* be emitted back to the application’s handlers:
* 1-)The grpc client with the orderer service uses a “regular” stateless HTTP connection in a request/response fashion with the “broadcast” call.
* The method implementation should emit “transaction submitted” when a successful acknowledgement is received in the response,
* or “error” when an error is received
* 2-)The method implementation should also maintain a persistent connection with the Chain’s event source Peer as part of the
* internal event hub mechanism in order to support the fabric events “BLOCK”, “CHAINCODE” and “TRANSACTION”.
* These events should cause the method to emit “complete” or “error” events to the application.
 */
func (c *channel) SendTransaction(tx *api.Transaction) ([]*api.TransactionResponse, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers is nil")
	}
	if tx == nil || tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("Transaction is nil")
	}
	// the original header
	hdr, err := protos_utils.GetHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protos_utils.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	envelope, err := c.SignPayload(paylBytes)
	if err != nil {
		return nil, err
	}

	transactionResponses, err := c.BroadcastEnvelope(envelope)
	if err != nil {
		return nil, err
	}

	return transactionResponses, nil
}

// SendInstantiateProposal ...
/**
* Sends an instantiate proposal to one or more endorsing peers.
* @param {string} chaincodeName: required - The name of the chain.
* @param {string} channelID: required - string of the name of the channel
* @param {[]string} args: optional - string Array arguments specific to the chaincode being instantiated
* @param {[]string} chaincodePath: required - string of the path to the location of the source code of the chaincode
* @param {[]string} chaincodeVersion: required - string of the version of the chaincode
 */
func (c *channel) SendInstantiateProposal(chaincodeName string, channelID string,
	args []string, chaincodePath string, chaincodeVersion string, targets []api.Peer) ([]*api.TransactionProposalResponse, string, error) {

	if chaincodeName == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeName' parameter")
	}
	if channelID == "" {
		return nil, "", fmt.Errorf("Missing 'channelID' parameter")
	}
	if chaincodePath == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodePath' parameter")
	}
	if chaincodeVersion == "" {

		return nil, "", fmt.Errorf("Missing 'chaincodeVersion' parameter")
	}

	targetPeers, err := c.getTargetPeers(targets)
	if err != nil {
		return nil, "", fmt.Errorf("GetTargetPeers return error: %s", err)
	}

	if len(targetPeers) < 1 {
		return nil, "", fmt.Errorf("Missing peer objects for instantiate CC proposal")
	}

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return nil, "", fmt.Errorf("Error getting creator: %v", err)
	}
	chaincodePolicy, err := buildChaincodePolicy(c.GetClientContext().GetConfig().GetFabricCAID())
	if err != nil {
		return nil, "", err
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, "", err
	}
	// create a proposal from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateDeployProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
	if err != nil {
		return nil, "", fmt.Errorf("Could not create chaincode Deploy proposal, err %s", err)
	}

	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return nil, "", err
	}

	transactionProposalResponse, err := c.SendTransactionProposal(&api.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TransactionID:  txID,
	}, 0, targetPeers)

	return transactionProposalResponse, txID, err
}

func (c *channel) SignPayload(payload []byte) (*api.SignedEnvelope, error) {
	//Get user info
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}

	signature, err := fc.SignObjectWithKey(payload, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}
	// here's the envelope
	return &api.SignedEnvelope{Payload: payload, Signature: signature}, nil
}

//broadcastEnvelope will send the given envelope to each orderer
func (c *channel) BroadcastEnvelope(envelope *api.SignedEnvelope) ([]*api.TransactionResponse, error) {
	// Check if orderers are defined
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	var responseMtx sync.Mutex
	var transactionResponses []*api.TransactionResponse
	var wg sync.WaitGroup

	for _, o := range c.orderers {
		wg.Add(1)
		go func(orderer api.Orderer) {
			defer wg.Done()
			var transactionResponse *api.TransactionResponse

			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.GetURL())
			if _, err := orderer.SendBroadcast(envelope); err != nil {
				logger.Debugf("Receive Error Response from orderer :%v\n", err)
				transactionResponse = &api.TransactionResponse{Orderer: orderer.GetURL(),
					Err: fmt.Errorf("Error calling orderer '%s':  %s", orderer.GetURL(), err)}
			} else {
				logger.Debugf("Receive Success Response from orderer\n")
				transactionResponse = &api.TransactionResponse{Orderer: orderer.GetURL(), Err: nil}
			}

			responseMtx.Lock()
			transactionResponses = append(transactionResponses, transactionResponse)
			responseMtx.Unlock()
		}(o)
	}
	wg.Wait()

	return transactionResponses, nil
}

// SendEnvelope sends the given envelope to each orderer and returns a block response
func (c *channel) SendEnvelope(envelope *api.SignedEnvelope) (*common.Block, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	var blockResponse *common.Block
	var errorResponse error
	var mutex sync.Mutex
	outstandingRequests := len(c.orderers)
	done := make(chan bool)

	// Send the request to all orderers and return as soon as one responds with a block.
	for _, o := range c.orderers {
		go func(orderer api.Orderer) {
			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.GetURL())
			blocks, errors := orderer.SendDeliver(envelope)
			select {
			case block := <-blocks:
				mutex.Lock()
				if blockResponse == nil {
					blockResponse = block
					done <- true
				}
				mutex.Unlock()

			case err := <-errors:
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = err
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()

			case <-time.After(time.Second * 5):
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = fmt.Errorf("Timeout waiting for response from orderer")
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()
			}
		}(o)
	}

	<-done

	if blockResponse != nil {
		return blockResponse, nil
	}

	// There must be an error
	if errorResponse != nil {
		return nil, fmt.Errorf("error returned from orderer service: %v", errorResponse)
	}

	return nil, fmt.Errorf("unexpected: didn't receive a block from any of the orderer servces and didn't receive any error")
}

func (c *channel) signProposal(proposal *pb.Proposal) (*pb.SignedProposal, error) {
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("Error mashalling proposal: %s", err)
	}

	signature, err := fc.SignObjectWithKey(proposalBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return nil, fmt.Errorf("Error signing proposal: %s", err)
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}

// fetchGenesisBlock fetches the configuration block for this channel
func (c *channel) fetchGenesisBlock() (*common.Block, error) {
	// Get user enrolment info and serialize for signing requests
	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("Error getting creator: %v", err)
	}
	// Seek block zero (the configuration tx for this channel)
	payload := fc.CreateSeekGenesisBlockRequest(c.name, creator)
	blockRequest, err := c.SignPayload(payload)
	if err != nil {
		return nil, fmt.Errorf("Error signing payload: %s", err)
	}
	// Request genesis block from ordering service
	block, err := c.SendEnvelope(blockRequest)
	if err != nil {
		return nil, fmt.Errorf("Error from SendEnvelope: %s", err.Error())
	}
	return block, nil
}

// internal utility method to build chaincode policy
// FIXME: for now always construct a 'Signed By any member of an organization by mspid' policy
func buildChaincodePolicy(mspid string) (*common.SignaturePolicyEnvelope, error) {
	// Define MSPRole
	memberRole, err := proto.Marshal(&mspprotos.MSPRole{Role: mspprotos.MSPRole_MEMBER, MspIdentifier: mspid})
	if err != nil {
		return nil, fmt.Errorf("Error marshal MSPRole: %s", err)
	}

	// construct a list of msp principals to select from using the 'n out of' operator
	onePrn := &mspprotos.MSPPrincipal{
		PrincipalClassification: mspprotos.MSPPrincipal_ROLE,
		Principal:               memberRole}

	// construct 'signed by msp principal at index 0'
	signedBy := &common.SignaturePolicy{Type: &common.SignaturePolicy_SignedBy{SignedBy: 0}}

	// construct 'one of one' policy
	oneOfone := &common.SignaturePolicy{Type: &common.SignaturePolicy_NOutOf_{NOutOf: &common.SignaturePolicy_NOutOf{
		N: 1, Rules: []*common.SignaturePolicy{signedBy}}}}

	p := &common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       oneOfone,
		Identities: []*mspprotos.MSPPrincipal{onePrn},
	}
	return p, nil
}

func loadConfigGroup(configItems *configItems, versionsGroup *common.ConfigGroup, group *common.ConfigGroup, name string, org string, top bool) error {
	logger.Debugf("loadConfigGroup - %s - START groups Org: %s", name, org)
	if group == nil {
		return nil
	}

	logger.Debugf("loadConfigGroup - %s   - version %v", name, group.Version)
	logger.Debugf("loadConfigGroup - %s   - mod policy %s", name, group.ModPolicy)
	logger.Debugf("loadConfigGroup - %s - >> groups", name)

	groups := group.GetGroups()
	if groups != nil {
		versionsGroup.Groups = make(map[string]*common.ConfigGroup)
		for key, configGroup := range groups {
			logger.Debugf("loadConfigGroup - %s - found config group ==> %s", name, key)
			// The Application group is where config settings are that we want to find
			versionsGroup.Groups[key] = &common.ConfigGroup{}
			loadConfigGroup(configItems, versionsGroup.Groups[key], configGroup, name+"."+key, key, false)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no groups", name)
	}
	logger.Debugf("loadConfigGroup - %s - << groups", name)

	logger.Debugf("loadConfigGroup - %s - >> values", name)
	values := group.GetValues()
	if values != nil {
		versionsGroup.Values = make(map[string]*common.ConfigValue)
		for key, configValue := range values {
			versionsGroup.Values[key] = &common.ConfigValue{}
			loadConfigValue(configItems, key, versionsGroup.Values[key], configValue, name, org)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no values", name)
	}
	logger.Debugf("loadConfigGroup - %s - << values", name)

	logger.Debugf("loadConfigGroup - %s - >> policies", name)
	policies := group.GetPolicies()
	if policies != nil {
		versionsGroup.Policies = make(map[string]*common.ConfigPolicy)
		for key, configPolicy := range policies {
			versionsGroup.Policies[key] = &common.ConfigPolicy{}
			loadConfigPolicy(configItems, key, versionsGroup.Policies[key], configPolicy, name, org)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no policies", name)
	}
	logger.Debugf("loadConfigGroup - %s - << policies", name)
	logger.Debugf("loadConfigGroup - %s - < group", name)
	return nil
}

func loadConfigValue(configItems *configItems, key string, versionsValue *common.ConfigValue, configValue *common.ConfigValue, groupName string, org string) error {
	logger.Infof("loadConfigValue - %s - START value name: %s", groupName, key)
	logger.Infof("loadConfigValue - %s   - version: %d", groupName, configValue.Version)
	logger.Infof("loadConfigValue - %s   - modPolicy: %s", groupName, configValue.ModPolicy)

	versionsValue.Version = configValue.Version

	switch key {
	case fabric_config.AnchorPeersKey:
		anchorPeers := &pb.AnchorPeers{}
		err := proto.Unmarshal(configValue.Value, anchorPeers)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal anchor peers from config value: %v", err)
		}

		logger.Debugf("loadConfigValue - %s   - AnchorPeers :: %s", groupName, anchorPeers)

		if len(anchorPeers.AnchorPeers) > 0 {
			for _, anchorPeer := range anchorPeers.AnchorPeers {
				oap := &api.OrgAnchorPeer{Org: org, Host: anchorPeer.Host, Port: anchorPeer.Port}
				configItems.anchorPeers = append(configItems.anchorPeers, oap)
				logger.Debugf("loadConfigValue - %s   - AnchorPeer :: %s:%d:%s", groupName, oap.Host, oap.Port, oap.Org)
			}
		}
		break

	case fabric_config.MSPKey:
		mspConfig := &mb.MSPConfig{}
		err := proto.Unmarshal(configValue.Value, mspConfig)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal MSPConfig from config value: %v", err)
		}

		logger.Debugf("loadConfigValue - %s   - MSP found", groupName)

		mspType := msp.ProviderType(mspConfig.Type)
		if mspType != msp.FABRIC {
			return fmt.Errorf("unsupported MSP type: %v", mspType)
		}

		configItems.msps = append(configItems.msps, mspConfig)
		break

	case fabric_config.ConsensusTypeKey:
		consensusType := &ab.ConsensusType{}
		err := proto.Unmarshal(configValue.Value, consensusType)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal ConsensusType from config value: %v", err)
		}

		logger.Debugf("loadConfigValue - %s   - Consensus type value :: %s", groupName, consensusType.Type)
		// TODO: Do something with this value
		break

	case fabric_config.BatchSizeKey:
		batchSize := &ab.BatchSize{}
		err := proto.Unmarshal(configValue.Value, batchSize)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal BatchSize from config value: %v", err)
		}

		logger.Debugf("loadConfigValue - %s   - BatchSize  maxMessageCount :: %d", groupName, batchSize.MaxMessageCount)
		logger.Debugf("loadConfigValue - %s   - BatchSize  absoluteMaxBytes :: %d", groupName, batchSize.AbsoluteMaxBytes)
		logger.Debugf("loadConfigValue - %s   - BatchSize  preferredMaxBytes :: %d", groupName, batchSize.PreferredMaxBytes)
		// TODO: Do something with this value
		break

	case fabric_config.BatchTimeoutKey:
		batchTimeout := &ab.BatchTimeout{}
		err := proto.Unmarshal(configValue.Value, batchTimeout)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal BatchTimeout from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - BatchTimeout timeout value :: %s", groupName, batchTimeout.Timeout)
		// TODO: Do something with this value
		break

	case fabric_config.ChannelRestrictionsKey:
		channelRestrictions := &ab.ChannelRestrictions{}
		err := proto.Unmarshal(configValue.Value, channelRestrictions)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal ChannelRestrictions from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - ChannelRestrictions max_count value :: %d", groupName, channelRestrictions.MaxCount)
		// TODO: Do something with this value
		break

	case fabric_config.HashingAlgorithmKey:
		hashingAlgorithm := &common.HashingAlgorithm{}
		err := proto.Unmarshal(configValue.Value, hashingAlgorithm)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal HashingAlgorithm from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - HashingAlgorithm names value :: %s", groupName, hashingAlgorithm.Name)
		// TODO: Do something with this value
		break

	case fabric_config.ConsortiumKey:
		consortium := &common.Consortium{}
		err := proto.Unmarshal(configValue.Value, consortium)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal Consortium from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - Consortium names value :: %s", groupName, consortium.Name)
		// TODO: Do something with this value
		break

	case fabric_config.BlockDataHashingStructureKey:
		bdhstruct := &common.BlockDataHashingStructure{}
		err := proto.Unmarshal(configValue.Value, bdhstruct)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal BlockDataHashingStructure from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - BlockDataHashingStructure width value :: %s", groupName, bdhstruct.Width)
		// TODO: Do something with this value
		break

	case fabric_config.OrdererAddressesKey:
		ordererAddresses := &common.OrdererAddresses{}
		err := proto.Unmarshal(configValue.Value, ordererAddresses)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal OrdererAddresses from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - OrdererAddresses addresses value :: %s", groupName, ordererAddresses.Addresses)
		if len(ordererAddresses.Addresses) > 0 {
			for _, ordererAddress := range ordererAddresses.Addresses {
				configItems.orderers = append(configItems.orderers, ordererAddress)
			}
		}
		break

	default:
		logger.Debugf("loadConfigValue - %s   - value: %s", groupName, configValue.Value)
	}
	return nil
}

func loadConfigPolicy(configItems *configItems, key string, versionsPolicy *common.ConfigPolicy, configPolicy *common.ConfigPolicy, groupName string, org string) error {
	logger.Debugf("loadConfigPolicy - %s - name: %s", groupName, key)
	logger.Debugf("loadConfigPolicy - %s - version: %d", groupName, configPolicy.Version)
	logger.Debugf("loadConfigPolicy - %s - mod_policy: %s", groupName, configPolicy.ModPolicy)

	versionsPolicy.Version = configPolicy.Version
	return loadPolicy(configItems, versionsPolicy, key, configPolicy.Policy, groupName, org)
}

func loadPolicy(configItems *configItems, versionsPolicy *common.ConfigPolicy, key string, policy *common.Policy, groupName string, org string) error {

	policyType := common.Policy_PolicyType(policy.Type)

	switch policyType {
	case common.Policy_SIGNATURE:
		sigPolicyEnv := &common.SignaturePolicyEnvelope{}
		err := proto.Unmarshal(policy.Value, sigPolicyEnv)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal SignaturePolicyEnvelope from config policy: %v", err)
		}
		logger.Debugf("loadConfigPolicy - %s - policy SIGNATURE :: %v", groupName, sigPolicyEnv.Rule)
		// TODO: Do something with this value
		break

	case common.Policy_MSP:
		// TODO: Not implemented yet
		logger.Debugf("loadConfigPolicy - %s - policy :: MSP POLICY NOT PARSED ", groupName)
		break

	case common.Policy_IMPLICIT_META:
		implicitMetaPolicy := &common.ImplicitMetaPolicy{}
		err := proto.Unmarshal(policy.Value, implicitMetaPolicy)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal ImplicitMetaPolicy from config policy: %v", err)
		}
		logger.Debugf("loadConfigPolicy - %s - policy IMPLICIT_META :: %v", groupName, implicitMetaPolicy)
		// TODO: Do something with this value
		break

	default:
		return fmt.Errorf("Unknown Policy type: %v", policyType)
	}
	return nil
}

// getBlock retrieves the block at the given position
func (c *channel) getBlock(pos *ab.SeekPosition) (*common.Block, error) {
	nonce, err := fc.GenerateRandomNonce()
	if err != nil {
		return nil, fmt.Errorf("error when generating nonce: %v", err)
	}

	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("error when serializing identity: %v", err)
	}

	txID, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, fmt.Errorf("error when generating TX ID: %v", err)
	}

	seekInfoHeader, err := BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, c.GetName(), txID, 0, "", time.Now())
	if err != nil {
		return nil, fmt.Errorf("error when building channel header: %v", err)
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, fmt.Errorf("error when marshalling channel header: %v", err)
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}

	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("error when marshalling signature header: %v", err)
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
		return nil, fmt.Errorf("error when marshalling seek info: %v", err)
	}

	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	seekPayloadBytes, err := proto.Marshal(seekPayload)
	if err != nil {
		return nil, err
	}

	signedEnvelope, err := c.SignPayload(seekPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("error when signing payload: %v", err)
	}

	return c.SendEnvelope(signedEnvelope)
}

func (c *channel) initializeFromConfig(configItems *configItems) error {
	// TODO revisit this if
	if len(configItems.msps) > 0 {
		msps, err := c.loadMSPs(configItems.msps)
		if err != nil {
			return fmt.Errorf("unable to load MSPs from config: %v", err)
		}

		if err := c.mspManager.Setup(msps); err != nil {
			return fmt.Errorf("error calling Setup on MSPManager: %v", err)
		}
	}
	c.anchorPeers = configItems.anchorPeers

	// TODO should we create orderers and endorsing peers
	return nil
}

func (c *channel) loadMSPs(mspConfigs []*mb.MSPConfig) ([]msp.MSP, error) {
	logger.Debugf("loadMSPs - start number of msps=%d", len(mspConfigs))

	msps := []msp.MSP{}
	for _, config := range mspConfigs {
		mspType := msp.ProviderType(config.Type)
		if mspType != msp.FABRIC {
			return nil, fmt.Errorf("MSP Configuration object type not supported: %v", mspType)
		}
		if len(config.Config) == 0 {
			return nil, fmt.Errorf("MSP Configuration object missing the payload in the 'Config' property")
		}

		fabricConfig := &mb.FabricMSPConfig{}
		err := proto.Unmarshal(config.Config, fabricConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to unmarshal FabricMSPConfig from config value: %v", err)
		}

		if fabricConfig.Name == "" {
			return nil, fmt.Errorf("MSP Configuration does not have a name")
		}

		// with this method we are only dealing with verifying MSPs, not local MSPs. Local MSPs are instantiated
		// from user enrollment materials (see User class). For verifying MSPs the root certificates are always
		// required
		if len(fabricConfig.RootCerts) == 0 {
			return nil, fmt.Errorf("MSP Configuration does not have any root certificates required for validating signing certificates")
		}

		// get the application org names
		var orgs []string
		orgUnits := fabricConfig.OrganizationalUnitIdentifiers
		for _, orgUnit := range orgUnits {
			logger.Debugf("loadMSPs - found org of :: %s", orgUnit.OrganizationalUnitIdentifier)
			orgs = append(orgs, orgUnit.OrganizationalUnitIdentifier)
		}

		// TODO: Do something with orgs

		newMSP, err := msp.NewBccspMsp()
		if err != nil {
			return nil, fmt.Errorf("error creating new MSP: %v", err)
		}

		if err := newMSP.Setup(config); err != nil {
			return nil, fmt.Errorf("error in Setup of new MSP: %v", err)
		}

		mspID, _ := newMSP.GetIdentifier()
		logger.Debugf("loadMSPs - adding msp=%s", mspID)

		msps = append(msps, newMSP)
	}

	logger.Debugf("loadMSPs - loaded %d MSPs", len(msps))
	return msps, nil
}

// BuildChannelHeader is a utility method to build a common chain header
func BuildChannelHeader(headerType common.HeaderType, channelID string, txID string, epoch uint64, chaincodeID string, timestamp time.Time) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: % chaincodeID: %s timestamp: %v", headerType, channelID, txID, epoch, chaincodeID, timestamp)
	channelHeader := &common.ChannelHeader{
		Type:      int32(headerType),
		Version:   1,
		ChannelId: channelID,
		TxId:      txID,
		Epoch:     epoch,
	}
	if !timestamp.IsZero() {
		ts := &proto_ts.Timestamp{
			Seconds: int64(timestamp.Second()),
			Nanos:   int32(timestamp.Nanosecond()),
		}
		channelHeader.Timestamp = ts
	}
	if chaincodeID != "" {
		ccID := &pb.ChaincodeID{
			Name: chaincodeID,
		}
		headerExt := &pb.ChaincodeHeaderExtension{
			ChaincodeId: ccID,
		}
		headerExtBytes, err := proto.Marshal(headerExt)
		if err != nil {
			return nil, fmt.Errorf("Error marshaling header extension: %v", err)
		}
		channelHeader.Extension = headerExtBytes
	}
	return channelHeader, nil
}
