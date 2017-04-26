/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fabricclient

import (
	"fmt"
	"strconv"
	"sync"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/hyperledger/fabric/bccsp"
	msp "github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"

	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"

	config "github.com/hyperledger/fabric-sdk-go/config"
	fabric_config "github.com/hyperledger/fabric/common/config"
	"github.com/hyperledger/fabric/common/configtx"
	mb "github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Chain ...
/**
 * The “Chain” object captures settings for a channel, which is created by
 * the orderers to isolate transactions delivery to peers participating on channel.
 * A chain must be initialized after it has been configured with the list of peers
 * and orderers. The initialization sends a CONFIGURATION transaction to the orderers
 * to create the specified channel and asks the peers to join that channel.
 *
 */
type Chain interface {
	GetName() string
	Initialize(data []byte) error
	IsSecurityEnabled() bool
	GetTCertBatchSize() int
	SetTCertBatchSize(batchSize int)
	AddPeer(peer Peer)
	RemovePeer(peer Peer)
	GetPeers() []Peer
	SetPrimaryPeer(peer Peer) error
	GetPrimaryPeer() Peer
	AddOrderer(orderer Orderer)
	RemoveOrderer(orderer Orderer)
	GetOrderers() []Orderer
	SetMSPManager(mspManager msp.MSPManager)
	GetMSPManager() msp.MSPManager
	JoinChannel(request *JoinChannelRequest) error
	UpdateChain() bool
	IsReadonly() bool
	QueryInfo() (*common.BlockchainInfo, error)
	QueryBlock(blockNumber int) (*common.Block, error)
	QueryBlockByHash(blockHash []byte) (*common.Block, error)
	QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error)
	QueryByChaincode(chaincodeName string, args []string, targets []Peer) ([][]byte, error)
	CreateTransactionProposal(chaincodeName string, chainID string, args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error)
	SendTransactionProposal(proposal *TransactionProposal, retry int, targets []Peer) ([]*TransactionProposalResponse, error)
	CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error)
	SendTransaction(tx *Transaction) ([]*TransactionResponse, error)
	SendInstantiateProposal(chaincodeName string, chainID string, args []string, chaincodePath string, chaincodeVersion string, targets []Peer) ([]*TransactionProposalResponse, string, error)
	GetOrganizationUnits() ([]string, error)
	QueryExtensionInterface() ChainExtension
}

// The ChainExtension interface allows extensions of the SDK to add functionality to Chain overloads.
type ChainExtension interface {
	GetClientContext() Client

	SignPayload(payload []byte) (*SignedEnvelope, error)
	BroadcastEnvelope(envelope *SignedEnvelope) ([]*TransactionResponse, error)

	// TODO: This should go somewhere else - see TransactionProposal.GetBytes(). - deprecated
	GetProposalBytes(tp *TransactionProposal) ([]byte, error)
}

type chain struct {
	name            string // Name of the chain is only meaningful to the client
	securityEnabled bool   // Security enabled flag
	peers           map[string]Peer
	tcertBatchSize  int // The number of tcerts to get in each batch
	orderers        map[string]Orderer
	clientContext   Client
	primaryPeer     Peer
	mspManager      msp.MSPManager
	anchorPeers     []*orgAnchorPeer
}

// The TransactionProposal object to be send to the endorsers
type TransactionProposal struct {
	TransactionID string

	signedProposal *pb.SignedProposal
	proposal       *pb.Proposal
}

// GetBytes returns the serialized bytes of this proposal
func (tp *TransactionProposal) GetBytes() ([]byte, error) {
	return proto.Marshal(tp.signedProposal)
}

// TransactionProposalResponse ...
/**
 * The TransactionProposalResponse result object returned from endorsers.
 */
type TransactionProposalResponse struct {
	Endorser string
	Err      error
	Status   int32

	proposal         *TransactionProposal
	proposalResponse *pb.ProposalResponse
}

// GetResponsePayload returns the response object payload
func (tpr *TransactionProposalResponse) GetResponsePayload() []byte {
	if tpr == nil || tpr.proposalResponse == nil {
		return nil
	}
	return tpr.proposalResponse.GetResponse().Payload
}

// GetPayload returns the response payload
func (tpr *TransactionProposalResponse) GetPayload() []byte {
	if tpr == nil || tpr.proposalResponse == nil {
		return nil
	}
	return tpr.proposalResponse.Payload
}

// The Transaction object created from an endorsed proposal
type Transaction struct {
	proposal    *TransactionProposal
	transaction *pb.Transaction
}

// TransactionResponse ...
/**
 * The TransactionProposalResponse result object returned from orderers.
 */
type TransactionResponse struct {
	Orderer string
	Err     error
}

// A SignedEnvelope can can be sent to an orderer for broadcasting
type SignedEnvelope struct {
	Payload   []byte
	signature []byte
}

// JoinChannelRequest allows a set of peers to transact on a channel on the network
type JoinChannelRequest struct {
	Targets []Peer
	TxID    string
	Nonce   []byte
}

// configItems contains the configuration values retrieved from the Orderer Service
type configItems struct {
	msps        []*mb.MSPConfig
	anchorPeers []*orgAnchorPeer
	orderers    []string
}

// orgAnchorPeer contains inormation about a peer
type orgAnchorPeer struct {
	org  string
	host string
	port int32
}

// NewChain ...
/**
 * @param {string} name to identify different chain instances. The naming of chain instances
 * is enforced by the ordering service and must be unique within the blockchain network
 * @param {Client} clientContext An instance of {@link Client} that provides operational context
 * such as submitting User etc.
 */
func NewChain(name string, client Client) (Chain, error) {
	if name == "" {
		return nil, fmt.Errorf("failed to create Chain. Missing required 'name' parameter")
	}
	if client == nil {
		return nil, fmt.Errorf("failed to create Chain. Missing required 'clientContext' parameter")
	}
	p := make(map[string]Peer)
	o := make(map[string]Orderer)
	c := &chain{name: name, securityEnabled: config.IsSecurityEnabled(), peers: p,
		tcertBatchSize: config.TcertBatchSize(), orderers: o, clientContext: client, mspManager: msp.NewMSPManager()}
	logger.Infof("Constructed Chain instance: %v", c)

	return c, nil
}

func (c *chain) QueryExtensionInterface() ChainExtension {
	return c
}

// GetClientContext returns the Client that was passed in to NewChain
func (c *chain) GetClientContext() Client {
	return c.clientContext
}

// GetProposalBytes returns the serialized transaction.
func (c *chain) GetProposalBytes(tp *TransactionProposal) ([]byte, error) {
	return proto.Marshal(tp.signedProposal)
}

// GetName ...
/**
 * Get the chain name.
 * @returns {string} The name of the chain.
 */
func (c *chain) GetName() string {
	return c.name
}

// IsSecurityEnabled ...
/**
 * Determine if security is enabled.
 */
func (c *chain) IsSecurityEnabled() bool {
	return c.securityEnabled
}

// GetTCertBatchSize ...
/**
 * Get the tcert batch size.
 */
func (c *chain) GetTCertBatchSize() int {
	return c.tcertBatchSize
}

// SetTCertBatchSize ...
/**
 * Set the tcert batch size.
 */
func (c *chain) SetTCertBatchSize(batchSize int) {
	c.tcertBatchSize = batchSize
}

// AddPeer ...
/**
 * Add peer endpoint to chain.
 * @param {Peer} peer An instance of the Peer that has been initialized with URL,
 * TLC certificate, and enrollment certificate.
 */
func (c *chain) AddPeer(peer Peer) {
	c.peers[peer.GetURL()] = peer
}

// RemovePeer ...
/**
 * Remove peer endpoint from chain.
 * @param {Peer} peer An instance of the Peer.
 */
func (c *chain) RemovePeer(peer Peer) {
	delete(c.peers, peer.GetURL())
}

// GetPeers ...
/**
 * Get peers of a chain from local information.
 * @returns {[]Peer} The peer list on the chain.
 */
func (c *chain) GetPeers() []Peer {
	var peersArray []Peer
	for _, v := range c.peers {
		peersArray = append(peersArray, v)
	}
	return peersArray
}

/**
 * Utility function to get target peers (target peer is valid only if it belongs to chain's peer list).
 * If targets is empty return chain's peer list
 * @returns {[]Peer} The target peer list
 * @returns {error} if target peer is not in chain's peer list
 */
func (c *chain) getTargetPeers(targets []Peer) ([]Peer, error) {

	if targets == nil || len(targets) == 0 {
		return c.GetPeers(), nil
	}

	var targetPeers []Peer
	for _, target := range targets {
		if !c.isValidPeer(target) {
			return nil, fmt.Errorf("The target peer must be on this chain peer list")
		}
		targetPeers = append(targetPeers, c.peers[target.GetURL()])
	}

	return targetPeers, nil
}

/**
 * Utility function to ensure that a peer exists on this chain
 * @returns {bool} true if peer exists on this chain
 */
func (c *chain) isValidPeer(peer Peer) bool {
	return peer != nil && c.peers[peer.GetURL()] != nil
}

// SetPrimaryPeer ...
/**
* Set the primary peer
* The peer to use for doing queries.
* Peer must be a peer on this chain's peer list.
* Default: When no primary peer has been set the first peer
* on the list will be used.
* @param {Peer} peer An instance of the Peer class.
* @returns error when peer is not on the existing peer list
 */
func (c *chain) SetPrimaryPeer(peer Peer) error {

	if !c.isValidPeer(peer) {
		return fmt.Errorf("The primary peer must be on this chain peer list")
	}

	c.primaryPeer = c.peers[peer.GetURL()]
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
func (c *chain) GetPrimaryPeer() Peer {

	if c.primaryPeer != nil {
		return c.primaryPeer
	}

	// When no primary peer has been set default to the first peer
	// from map range - order is not guaranteed
	for _, peer := range c.peers {
		return peer
	}

	return nil
}

// AddOrderer ...
/**
 * Add orderer endpoint to a chain object, this is a local-only operation.
 * A chain instance may choose to use a single orderer node, which will broadcast
 * requests to the rest of the orderer network. Or if the application does not trust
 * the orderer nodes, it can choose to use more than one by adding them to the chain instance.
 * All APIs concerning the orderer will broadcast to all orderers simultaneously.
 * @param {Orderer} orderer An instance of the Orderer class.
 */
func (c *chain) AddOrderer(orderer Orderer) {
	c.orderers[orderer.GetURL()] = orderer
}

// RemoveOrderer ...
/**
 * Remove orderer endpoint from a chain object, this is a local-only operation.
 * @param {Orderer} orderer An instance of the Orderer class.
 */
func (c *chain) RemoveOrderer(orderer Orderer) {
	delete(c.orderers, orderer.GetURL())

}

// GetOrderers ...
/**
 * Get orderers of a chain.
 */
func (c *chain) GetOrderers() []Orderer {
	var orderersArray []Orderer
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
func (c *chain) SetMSPManager(mspManager msp.MSPManager) {
	c.mspManager = mspManager
}

// GetMSPManager returns the MSP Manager for this channel
func (c *chain) GetMSPManager() msp.MSPManager {
	return c.mspManager
}

// GetOrganizationUnits - to get identifier for the organization configured on the channel
func (c *chain) GetOrganizationUnits() ([]string, error) {
	chainMSPManager := c.GetMSPManager()
	msps, err := chainMSPManager.GetMSPs()
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

// Initialize initializes the chain
/**
 * Retrieve the configuration from the orderer and initializes this chain (channel)
 * with those values. Optionally a configuration may be passed in to initialize this channel
 * without making the call to the orderer.
 * @param {byte[]} config - Optional - A serialized form of the protobuf configuration update envelope
 */
func (c *chain) Initialize(config []byte) error {
	var configItems *configItems

	if len(config) > 0 {
		var err error
		if configItems, err = c.loadConfigUpdateEnvelope(config); err != nil {
			return fmt.Errorf("Unable to load config update envelope: %v", err)
		}
	} else {
		configEnvelope, err := c.getChannelConfig()
		if err != nil {
			return fmt.Errorf("Unable to retrieve channel configuration from orderer service: %v", err)
		}

		configItems, err = c.loadConfigEnvelope(configEnvelope)
		if err != nil {
			return fmt.Errorf("Unable to load config envelope: %v", err)
		}
	}

	return c.initializeFromConfig(configItems)
}

// JoinChannel instructs a set of peers to join the channel represented by
// this chain
// @param {JoinChannelRequest} Join channel request
// @returns error, if applicable
func (c *chain) JoinChannel(request *JoinChannelRequest) error {
	joinCommand := "JoinChain"
	err := validateJoinChannelRequest(request)
	if err != nil {
		return err
	}
	// Fetch genesis block
	block, err := c.fetchGenesisBlock()
	if err != nil {
		return err
	}
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		return fmt.Errorf("Error unmarshalling block: %s", err)
	}
	// Get user enrolment info and serialize for signing requests
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}
	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return err
	}
	// Create join channel transaction proposal for target peers
	var args [][]byte
	args = append(args, []byte(joinCommand))
	args = append(args, blockBytes)
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input: &pb.ChaincodeInput{Args: args}}}

	// create a proposal from a ChaincodeInvocationSpec
	proposal, _, err := protos_utils.
		CreateChaincodeProposalWithTxIDNonceAndTransient(request.TxID,
			common.HeaderType_ENDORSER_TRANSACTION, "", ccis,
			request.Nonce, creatorID, nil)
	if err != nil {
		return fmt.Errorf("Could not create chaincode proposal, err %s", err)
	}
	// Serialize join proposal
	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return err
	}
	// Sign join proposal
	signature, err := signObjectWithKey(proposalBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return err
	}
	// Send join proposal
	proposalResponses, err := c.SendTransactionProposal(&TransactionProposal{
		TransactionID: request.TxID,
		signedProposal: &pb.SignedProposal{
			ProposalBytes: proposalBytes,
			Signature:     signature},
		proposal: proposal,
	}, 0, request.Targets)
	if err != nil {
		return fmt.Errorf("Error sending join transaction proposal: %s", err)
	}

	// Check responses from target peers for success/failure
	var joinError string
	for _, response := range proposalResponses {
		if response.Err != nil {
			joinError = joinError +
				fmt.Sprintf("Join channel proposal response error: %s \n",
					response.Err.Error())
		}
	}
	if joinError != "" {
		return fmt.Errorf(joinError)
	}

	return nil
}

// UpdateChain ...
/**
 * Calls the orderer(s) to update an existing chain. This allows the addition and
 * deletion of Peer nodes to an existing chain, as well as the update of Peer
 * certificate information upon certificate renewals.
 * @returns {bool} Whether the chain update process was successful.
 */
func (c *chain) UpdateChain() bool {
	return false
}

// IsReadonly ...
/**
 * Get chain status to see if the underlying channel has been terminated,
 * making it a read-only chain, where information (transactions and states)
 * can be queried but no new transactions can be submitted.
 * @returns {bool} Is read-only, true or not.
 */
func (c *chain) IsReadonly() bool {
	return false //to do
}

// QueryInfo ...
/**
 * Queries for various useful information on the state of the Chain
 * (height, known peers).
 * This query will be made to the primary peer.
 * @returns {object} With height, currently the only useful info.
 */
func (c *chain) QueryInfo() (*common.BlockchainInfo, error) {

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

// QueryBlock ...
/**
 * Queries the ledger for Block by block number.
 * This query will be made to the primary peer.
 * @param {int} blockNumber The number which is the ID of the Block.
 * @returns {object} Object containing the block.
 */
func (c *chain) QueryBlock(blockNumber int) (*common.Block, error) {

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

// QueryBlockByHash ...
/**
 * Queries the ledger for Block by block hash.
 * This query will be made to the primary peer.
 * @param {byte[]} block hash of the Block.
 * @returns {object} Object containing the block.
 */
func (c *chain) QueryBlockByHash(blockHash []byte) (*common.Block, error) {

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

// QueryTransaction ...
/**
* Queries the ledger for Transaction by number.
* This query will be made to the primary peer.
* @param {int} transactionID
* @returns {object} ProcessedTransaction information containing the transaction.
 */
func (c *chain) QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error) {

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
func (c *chain) QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error) {

	payload, err := c.queryByChaincodeByTarget("lccc", []string{"getchaincodes"}, c.GetPrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke lccc getchaincodes return error: %v", err)
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
func (c *chain) queryByChaincodeByTarget(chaincodeName string, args []string, target Peer) ([]byte, error) {

	queryResponses, err := c.QueryByChaincode(chaincodeName, args, []Peer{target})
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
 * @returns {[][]byte} an array of payloads
*/
func QueryByChaincode(chaincodeName string, args []string, targets []Peer, clientContext Client) ([][]byte, error) {
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
			responses = append(responses, response.GetResponsePayload())
		}
	}

	if len(errMsg) > 0 {
		return responses, fmt.Errorf(errMsg)
	}

	return responses, nil
}

func (c *chain) QueryByChaincode(chaincodeName string, args []string, targets []Peer) ([][]byte, error) {
	return QueryByChaincode(chaincodeName, args, targets, c.clientContext)
}

// CreateTransactionProposal ...
/**
 * Create  a proposal for transaction. This involves assembling the proposal
 * with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
 * ECert to sign.
 */
func (c *chain) CreateTransactionProposal(chaincodeName string, chainID string,
	args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error) {
	return CreateTransactionProposal(chaincodeName, chainID, args, sign, transientData, c.clientContext)
}

//CreateTransactionProposal  ...
func CreateTransactionProposal(chaincodeName string, chainID string,
	args []string, sign bool, transientData map[string][]byte, clientContext Client) (*TransactionProposal, error) {

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	user, err := clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, err
	}
	// create a proposal from a ChaincodeInvocationSpec
	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTransient(common.HeaderType_ENDORSER_TRANSACTION, chainID, ccis, creatorID, transientData)
	if err != nil {
		return nil, fmt.Errorf("Could not create chaincode proposal, err %s", err)
	}

	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, err
	}

	signature, err := signObjectWithKey(proposalBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil, clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}
	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	return &TransactionProposal{
		TransactionID:  txID,
		signedProposal: signedProposal,
		proposal:       proposal,
	}, nil

}

// SendTransactionProposal ...
// Send  the created proposal to peer for endorsement.
func (c *chain) SendTransactionProposal(proposal *TransactionProposal, retry int, targets []Peer) ([]*TransactionProposalResponse, error) {
	if c.peers == nil || len(c.peers) == 0 {
		return nil, fmt.Errorf("peers is nil")
	}
	if proposal == nil || proposal.signedProposal == nil {
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
func SendTransactionProposal(proposal *TransactionProposal, retry int, targetPeers []Peer) ([]*TransactionProposalResponse, error) {

	if proposal == nil || proposal.signedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	if len(targetPeers) < 1 {
		return nil, fmt.Errorf("Missing peer objects for sending transaction proposal")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range targetPeers {
		wg.Add(1)
		go func(peer Peer) {
			defer wg.Done()
			var err error
			var proposalResponse *TransactionProposalResponse
			logger.Debugf("Send ProposalRequest to peer :%s\n", peer.GetURL())
			if proposalResponse, err = peer.SendProposal(proposal); err != nil {
				logger.Debugf("Receive Error Response :%v\n", proposalResponse)
				proposalResponse = &TransactionProposalResponse{
					Endorser: peer.GetURL(),
					Err:      fmt.Errorf("Error calling endorser '%s':  %s", peer.GetURL(), err),
					proposal: proposal,
				}
			} else {
				prp1, _ := protos_utils.GetProposalResponsePayload(proposalResponse.proposalResponse.Payload)
				act1, _ := protos_utils.GetChaincodeAction(prp1.Extension)
				logger.Debugf("%s ProposalResponsePayload Extension ChaincodeAction Results\n%s\n", peer.GetURL(), string(act1.Results))

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
func (c *chain) CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("At least one proposal response is necessary")
	}

	proposal := resps[0].proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.proposal.Payload)
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

	//	var a1 []byte
	//	for n, r := range resps {
	//		if n == 0 {
	//			a1 = r.Payload
	//			if r.Response.Status != 200 {
	//				return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
	//			}
	//			continue
	//		}

	//		if bytes.Compare(a1, r.Payload) != 0 {
	//			return nil, fmt.Errorf("ProposalResponsePayloads do not match")
	//		}
	//	}

	for _, r := range resps {
		if r.proposalResponse.Response.Status != 200 {
			return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.proposalResponse.Response.Status, r.proposalResponse.Response.Message)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.proposalResponse.Endorsement
	}
	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].proposalResponse.Payload, Endorsements: endorsements}

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

	return &Transaction{
		transaction: &pb.Transaction{Actions: taas},
		proposal:    proposal,
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
func (c *chain) SendTransaction(tx *Transaction) ([]*TransactionResponse, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers is nil")
	}
	if tx == nil || tx.proposal == nil || tx.proposal.proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("Transaction is nil")
	}
	// the original header
	hdr, err := protos_utils.GetHeader(tx.proposal.proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx.transaction)
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
* @param {string} chainID: required - string of the name of the chain
* @param {[]string} args: optional - string Array arguments specific to the chaincode being instantiated
* @param {[]string} chaincodePath: required - string of the path to the location of the source code of the chaincode
* @param {[]string} chaincodeVersion: required - string of the version of the chaincode
 */
func (c *chain) SendInstantiateProposal(chaincodeName string, chainID string,
	args []string, chaincodePath string, chaincodeVersion string, targets []Peer) ([]*TransactionProposalResponse, string, error) {

	if chaincodeName == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeName' parameter")
	}
	if chainID == "" {
		return nil, "", fmt.Errorf("Missing 'chainID' parameter")
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

	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, "", fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, "", err
	}
	chaincodePolicy, err := buildChaincodePolicy(config.GetFabricCAID())
	if err != nil {
		return nil, "", err
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, "", err
	}
	// create a proposal from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateDeployProposalFromCDS(chainID, ccds, creatorID, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
	if err != nil {
		return nil, "", fmt.Errorf("Could not create chaincode Deploy proposal, err %s", err)
	}

	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return nil, "", err
	}

	transactionProposalResponse, err := c.SendTransactionProposal(&TransactionProposal{
		signedProposal: signedProposal,
		proposal:       proposal,
		TransactionID:  txID,
	}, 0, targetPeers)

	return transactionProposalResponse, txID, err
}

func (c *chain) SignPayload(payload []byte) (*SignedEnvelope, error) {
	//Get user info
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}

	signature, err := signObjectWithKey(payload, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}
	// here's the envelope
	return &SignedEnvelope{Payload: payload, signature: signature}, nil
}

//broadcastEnvelope will send the given envelope to each orderer
func (c *chain) BroadcastEnvelope(envelope *SignedEnvelope) ([]*TransactionResponse, error) {
	// Check if orderers are defined
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	var responseMtx sync.Mutex
	var transactionResponses []*TransactionResponse
	var wg sync.WaitGroup

	for _, o := range c.orderers {
		wg.Add(1)
		go func(orderer Orderer) {
			defer wg.Done()
			var transactionResponse *TransactionResponse

			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.GetURL())
			if err := orderer.SendBroadcast(envelope); err != nil {
				logger.Debugf("Receive Error Response from orderer :%v\n", err)
				transactionResponse = &TransactionResponse{orderer.GetURL(),
					fmt.Errorf("Error calling orderer '%s':  %s", orderer.GetURL(), err)}
			} else {
				logger.Debugf("Receive Success Response from orderer\n")
				transactionResponse = &TransactionResponse{orderer.GetURL(), nil}
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
func (c *chain) SendEnvelope(envelope *SignedEnvelope) (*common.Block, error) {
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
		go func(orderer Orderer) {
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

// signObjectWithKey will sign the given object with the given key,
// hashOpts and signerOpts
func signObjectWithKey(object []byte, key bccsp.Key,
	hashOpts bccsp.HashOpts, signerOpts bccsp.SignerOpts, cryptoSuite bccsp.BCCSP) ([]byte, error) {
	digest, err := cryptoSuite.Hash(object, hashOpts)
	if err != nil {
		return nil, err
	}
	signature, err := cryptoSuite.Sign(key, digest, signerOpts)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (c *chain) signProposal(proposal *pb.Proposal) (*pb.SignedProposal, error) {
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, err
	}

	signature, err := signObjectWithKey(proposalBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}

// fetchGenesisBlock fetches the configuration block for this channel
func (c *chain) fetchGenesisBlock() (*common.Block, error) {
	// Get user enrolment info and serialize for signing requests
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}
	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, err
	}
	// Seek block zero (the configuration tx for this channel)
	payload := util.CreateGenesisBlockRequest(c.name, creatorID)
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

func getSerializedIdentity(userCertificate []byte) ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{Mspid: config.GetFabricCAID(),
		IdBytes: userCertificate}
	creatorID, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, fmt.Errorf("Could not Marshal serializedIdentity, err %s", err)
	}
	return creatorID, nil
}

// internal utility method to build chaincode policy
// FIXME: for now always construct a 'Signed By any member of an organization by mspid' policy
func buildChaincodePolicy(mspid string) (*common.SignaturePolicyEnvelope, error) {
	// Define MSPRole
	memberRole, err := proto.Marshal(&common.MSPRole{Role: common.MSPRole_MEMBER, MspIdentifier: mspid})
	if err != nil {
		return nil, fmt.Errorf("Error marshal MSPRole: %s", err)
	}

	// construct a list of msp principals to select from using the 'n out of' operator
	onePrn := &common.MSPPrincipal{
		PrincipalClassification: common.MSPPrincipal_ROLE,
		Principal:               memberRole}

	// construct 'signed by msp principal at index 0'
	signedBy := &common.SignaturePolicy{Type: &common.SignaturePolicy_SignedBy{SignedBy: 0}}

	// construct 'one of one' policy
	oneOfone := &common.SignaturePolicy{Type: &common.SignaturePolicy_NOutOf_{NOutOf: &common.SignaturePolicy_NOutOf{
		N: 1, Policies: []*common.SignaturePolicy{signedBy}}}}

	p := &common.SignaturePolicyEnvelope{
		Version:    0,
		Policy:     oneOfone,
		Identities: []*common.MSPPrincipal{onePrn},
	}
	return p, nil
}

func validateJoinChannelRequest(request *JoinChannelRequest) error {
	// Validate arguments
	if request == nil {
		return fmt.Errorf("JoinChannelRequest argument is required to join channel")
	}
	if request.Targets == nil || len(request.Targets) == 0 {
		return fmt.Errorf("Atleast one target peer is required to join channel")
	}
	if request.TxID == "" {
		return fmt.Errorf("Transaction ID is required to join channel")
	}
	if request.Nonce == nil {
		return fmt.Errorf("Nonce is required to join channel")
	}
	return nil
}

func (c *chain) getChannelConfig() (*common.ConfigEnvelope, error) {
	// Get the newest block
	block, err := c.getBlock(util.NewNewestSeekPosition())
	if err != nil {
		return nil, err
	}

	logger.Debugf("GetChannelConfig - Retrieved newest block index: %d\n", block.Header.Number)

	// Get the index of the last config block
	lastConfig, err := util.GetLastConfigFromBlock(block)
	if err != nil {
		return nil, fmt.Errorf("Unable to get last config from block: %v", err)
	}

	logger.Debugf("GetChannelConfig - Last config index: %d\n", lastConfig.Index)

	// Get the last config block
	block, err = c.getBlock(util.NewSpecificSeekPosition(lastConfig.Index))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve block at index %d: %v", lastConfig.Index, err)
	}

	logger.Debugf("GetChannelConfig - Last config block number %d, Number of tx: %d", block.Header.Number, len(block.Data.Data))

	if len(block.Data.Data) != 1 {
		return nil, fmt.Errorf("Config block must only contain one transaction but contains %d", len(block.Data.Data))
	}

	envelope, err := protos_utils.ExtractEnvelope(block, 0)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract envelope from config block: %v", err)
	}

	payload, err := protos_utils.ExtractPayload(envelope)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract payload from config block envelope: %v", err)
	}

	channelHeader, err := protos_utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract channel header from config block payload: %v", err)
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

func (c *chain) loadConfigEnvelope(config *common.ConfigEnvelope) (*configItems, error) {
	configItems := &configItems{
		msps:        []*mb.MSPConfig{},
		anchorPeers: []*orgAnchorPeer{},
		orderers:    []string{},
	}
	err := loadConfigGroup(configItems, config.Config.ChannelGroup, "base", "", true, false)
	if err != nil {
		return nil, fmt.Errorf("Unable to load config items from channel group: %v", err)
	}

	logger.Debugf("Chain config: %v", configItems)

	return configItems, nil
}

func (c *chain) loadConfigUpdateEnvelope(data []byte) (*configItems, error) {
	logger.Debugf("loadConfigUpdateEnvelope - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(data, envelope)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal envelope: %v", err)
	}

	payload, err := protos_utils.ExtractPayload(envelope)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract payload from config update envelope: %v", err)
	}

	channelHeader, err := protos_utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract channel header from config update payload: %v", err)
	}

	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG_UPDATE {
		return nil, fmt.Errorf("Block must be of type 'CONFIG_UPDATE'")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	if err := proto.Unmarshal(payload.Data, configUpdateEnvelope); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal config update envelope: %v", err)
	}

	configUpdate := &common.ConfigUpdate{}
	if err := proto.Unmarshal(configUpdateEnvelope.ConfigUpdate, configUpdate); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal config update from envelope: %v", err)
	}

	logger.Debugf("loadConfigUpdateEnvelope - channel ::%s", configUpdate.ChannelId)

	configItems := &configItems{
		msps:        []*mb.MSPConfig{},
		anchorPeers: []*orgAnchorPeer{},
		orderers:    []string{},
	}

	// Do the write_set second so they update anything in the read set
	err = loadConfigGroup(configItems, configUpdate.ReadSet, "read_set", "", true, false)
	err = loadConfigGroup(configItems, configUpdate.WriteSet, "write_set", "", true, false)

	return configItems, nil
}

func loadConfigGroup(configItems *configItems, group *common.ConfigGroup, name string, org string, top bool, keepChildren bool) error {
	logger.Debugf("loadConfigGroup - %s - START groups Org: %s  keep: %v", name, org, keepChildren)
	if group == nil {
		return nil
	}

	groups := group.GetGroups()
	if groups != nil {
		for key, g := range groups {
			logger.Debugf("loadConfigGroup - %s - found config group ==> %s", name, key)
			// The Application group is where config settings are that we want to find
			loadConfigGroup(configItems, g, name+"."+key, key, false, (key == fabric_config.ApplicationGroupKey || keepChildren))
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no groups", name)
	}

	logger.Debugf("loadConfigGroup - %s - START values", name)

	values := group.GetValues()
	if values != nil {
		for key, configValue := range values {
			loadConfigValue(configItems, key, configValue, name, org, keepChildren)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no values", name)
	}

	logger.Debugf("loadConfigGroup - %s - END values", name)

	logger.Debugf("loadConfigGroup - %s - START policies", name)

	policies := group.GetPolicies()
	if policies != nil {
		for key, configPolicy := range policies {
			loadConfigPolicy(configItems, key, configPolicy, name, org, keepChildren)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no policies", name)
	}

	logger.Debugf("loadConfigGroup - %s - END policies", name)

	logger.Debugf("loadConfigGroup - %s - END group", name)

	return nil
}

func loadConfigValue(configItems *configItems, key string, configValue *common.ConfigValue, groupName string, org string, keepValue bool) error {
	logger.Debugf("loadConfigValue - %s - START value name: %s  keep:%v", groupName, key, keepValue)
	logger.Debugf("loadConfigValue - %s   - version: %d", groupName, configValue.Version)
	logger.Debugf("loadConfigValue - %s   - modPolicy: %s", groupName, configValue.ModPolicy)

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
				oap := &orgAnchorPeer{org: org, host: anchorPeer.Host, port: anchorPeer.Port}
				configItems.anchorPeers = append(configItems.anchorPeers, oap)
				logger.Debugf("loadConfigValue - %s   - AnchorPeer :: %s:%d:%s", groupName, oap.host, oap.port, oap.org)
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

		if keepValue {
			mspType := msp.ProviderType(mspConfig.Type)
			if mspType != msp.FABRIC {
				return fmt.Errorf("unsupported MSP type: %v", mspType)
			}

			configItems.msps = append(configItems.msps, mspConfig)
		}
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

	case configtx.CreationPolicyKey:
		creationPolicy := &ab.CreationPolicy{}
		err := proto.Unmarshal(configValue.Value, creationPolicy)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal CreationPolicy from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - CreationPolicy policy value :: %s", groupName, creationPolicy.Policy)
		// TODO: Do something with this value
		break

	case fabric_config.ChainCreationPolicyNamesKey:
		chainCreationPolicyNames := &ab.ChainCreationPolicyNames{}
		err := proto.Unmarshal(configValue.Value, chainCreationPolicyNames)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal ChainCreationPolicyNames from config value: %v", err)
		}
		logger.Debugf("loadConfigValue - %s   - ChainCreationPolicyNames names value :: %s", groupName, chainCreationPolicyNames.Names)
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

func loadConfigPolicy(configItems *configItems, key string, configPolicy *common.ConfigPolicy, groupName string, org string, keepPolicy bool) error {
	logger.Debugf("loadConfigPolicy - %s - name: %s  keep: %v", groupName, key, keepPolicy)
	logger.Debugf("loadConfigPolicy - %s - version: %d", groupName, configPolicy.Version)
	logger.Debugf("loadConfigPolicy - %s - mod_policy: %s", groupName, configPolicy.ModPolicy)

	policyType := common.Policy_PolicyType(configPolicy.Policy.Type)

	switch policyType {
	case common.Policy_SIGNATURE:
		sigPolicyEnv := &common.SignaturePolicyEnvelope{}
		err := proto.Unmarshal(configPolicy.Policy.Policy, sigPolicyEnv)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal SignaturePolicyEnvelope from config policy: %v", err)
		}
		logger.Debugf("loadConfigPolicy - %s - policy SIGNATURE :: %v", groupName, sigPolicyEnv.Policy)
		// TODO: Do something with this value
		break

	case common.Policy_MSP:
		// TODO: Not implemented yet
		logger.Debugf("loadConfigPolicy - %s - policy :: MSP POLICY NOT PARSED ", groupName)
		break

	case common.Policy_IMPLICIT_META:
		implicitMetaPolicy := &common.ImplicitMetaPolicy{}
		err := proto.Unmarshal(configPolicy.Policy.Policy, implicitMetaPolicy)
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
func (c *chain) getBlock(pos *ab.SeekPosition) (*common.Block, error) {
	nonce, err := util.GenerateRandomNonce()
	if err != nil {
		return nil, fmt.Errorf("error when generating nonce: %v", err)
	}

	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	creator, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, fmt.Errorf("error when serializing identity: %v", err)
	}

	txID, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, fmt.Errorf("error when generating TX ID: %v", err)
	}

	seekInfoHeader, err := util.BuildChannelHeader(c.GetName(), common.HeaderType_DELIVER_SEEK_INFO, txID, 0)
	if err != nil {
		return nil, fmt.Errorf("error when building channel header: %v", err)
	}

	seekInfoHeaderBytes, err := protos_utils.Marshal(seekInfoHeader)
	if err != nil {
		return nil, fmt.Errorf("error when marshalling channel header: %v", err)
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}

	signatureHeaderBytes, err := protos_utils.Marshal(signatureHeader)
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

	seekInfoBytes, err := protos_utils.Marshal(seekInfo)
	if err != nil {
		return nil, fmt.Errorf("error when marshalling seek info: %v", err)
	}

	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	seekPayloadBytes, err := protos_utils.Marshal(seekPayload)
	if err != nil {
		return nil, err
	}

	signedEnvelope, err := c.SignPayload(seekPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("error when signing payload: %v", err)
	}

	return c.SendEnvelope(signedEnvelope)
}

func (c *chain) initializeFromConfig(configItems *configItems) error {
	msps, err := c.loadMSPs(configItems.msps)
	if err != nil {
		return fmt.Errorf("unable to load MSPs from config: %v", err)
	}

	if err := c.mspManager.Setup(msps); err != nil {
		return fmt.Errorf("error calling Setup on MSPManager: %v", err)
	}

	c.anchorPeers = configItems.anchorPeers

	// TODO should we create orderers and endorsing peers
	return nil
}

func (c *chain) loadMSPs(mspConfigs []*mb.MSPConfig) ([]msp.MSP, error) {
	logger.Debugf("loadMSPs - start number of msps=%d", len(mspConfigs))

	var msps []msp.MSP
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
