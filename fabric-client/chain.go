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
	"sync"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/bccsp"
	msp "github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"

	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"

	config "github.com/hyperledger/fabric-sdk-go/config"
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
	IsSecurityEnabled() bool
	GetTCertBatchSize() int
	SetTCertBatchSize(batchSize int)
	AddPeer(peer Peer)
	RemovePeer(peer Peer)
	GetPeers() []Peer
	AddOrderer(orderer Orderer)
	RemoveOrderer(orderer Orderer)
	GetOrderers() []Orderer
	InitializeChain() bool
	UpdateChain() bool
	IsReadonly() bool
	QueryInfo()
	QueryBlock(blockNumber int)
	QueryTransaction(transactionID int)
	CreateTransactionProposal(chaincodeName string, chainID string, args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error)
	SendTransactionProposal(proposal *TransactionProposal, retry int) ([]*TransactionProposalResponse, error)
	CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error)
	SendTransaction(tx *Transaction) ([]*TransactionResponse, error)
	SendInstallProposal(chaincodeName string, chaincodePath string, chaincodeVersion string, chaincodePackage []byte) ([]*TransactionProposalResponse, string, error)
	SendInstantiateProposal(chaincodeName string, chainID string, args []string, chaincodePath string, chaincodeVersion string) ([]*TransactionProposalResponse, string, error)
}

type chain struct {
	name            string // Name of the chain is only meaningful to the client
	securityEnabled bool   // Security enabled flag
	peers           map[string]Peer
	tcertBatchSize  int // The number of tcerts to get in each batch
	orderers        map[string]Orderer
	clientContext   Client
}

// The TransactionProposal object to be send to the endorsers
type TransactionProposal struct {
	TransactionID string

	signedProposal *pb.SignedProposal
	proposal       *pb.Proposal
}

// TransactionProposalResponse ...
/**
 * The TransactionProposalResponse result object returned from endorsers.
 */
type TransactionProposalResponse struct {
	Endorser string
	Err      error

	proposal         *TransactionProposal
	proposalResponse *pb.ProposalResponse
}

// GetResponsePayload returns the response payload
func (tpr *TransactionProposalResponse) GetResponsePayload() []byte {
	if tpr == nil || tpr.proposalResponse == nil {
		return nil
	}
	return tpr.proposalResponse.GetResponse().Payload
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

// NewChain ...
/**
 * @param {string} name to identify different chain instances. The naming of chain instances
 * is enforced by the ordering service and must be unique within the blockchain network
 * @param {Client} clientContext An instance of {@link Client} that provides operational context
 * such as submitting User etc.
 */
func NewChain(name string, client Client) (Chain, error) {
	if name == "" {
		return nil, fmt.Errorf("Failed to create Chain. Missing requirement 'name' parameter.")
	}
	if client == nil {
		return nil, fmt.Errorf("Failed to create Chain. Missing requirement 'clientContext' parameter.")
	}
	p := make(map[string]Peer)
	o := make(map[string]Orderer)
	c := &chain{name: name, securityEnabled: config.IsSecurityEnabled(), peers: p,
		tcertBatchSize: config.TcertBatchSize(), orderers: o, clientContext: client}
	logger.Infof("Constructed Chain instance: %v", c)

	return c, nil
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

// InitializeChain ...
/**
 * Calls the orderer(s) to start building the new chain, which is a combination
 * of opening new message stream and connecting the list of participating peers.
 * This is a long-running process. Only one of the application instances needs
 * to call this method. Once the chain is successfully created, other application
 * instances only need to call getChain() to obtain the information about this chain.
 * @returns {bool} Whether the chain initialization process was successful.
 */
func (c *chain) InitializeChain() bool {
	return false
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
 * @returns {object} With height, currently the only useful info.
 */
func (c *chain) QueryInfo() {
	//to do
}

// QueryBlock ...
/**
 * Queries the ledger for Block by block number.
 * @param {int} blockNumber The number which is the ID of the Block.
 * @returns {object} Object containing the block.
 */
func (c *chain) QueryBlock(blockNumber int) {
	//to do
}

// QueryTransaction ...
/**
 * Queries the ledger for Transaction by number.
 * @param {int} transactionID
 * @returns {object} Transaction information containing the transaction.
 */
func (c *chain) QueryTransaction(transactionID int) {
	//to do
}

// CreateTransactionProposal ...
/**
 * Create  a proposal for transaction. This involves assembling the proposal
 * with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
 * ECert to sign.
 */
func (c *chain) CreateTransactionProposal(chaincodeName string, chainID string,
	args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error) {

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, fmt.Errorf("GetUserContext return error: %s", err)
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

	signature, err := c.signObjectWithKey(proposalBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil)
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
func (c *chain) SendTransactionProposal(proposal *TransactionProposal, retry int) ([]*TransactionProposalResponse, error) {
	if c.peers == nil || len(c.peers) == 0 {
		return nil, fmt.Errorf("peers is nil")
	}
	if proposal == nil || proposal.signedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range c.peers {
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
	envelope, err := c.signPayload(paylBytes)
	if err != nil {
		return nil, err
	}

	transactionResponses, err := c.broadcastEnvelope(envelope)
	if err != nil {
		return nil, err
	}

	return transactionResponses, nil
}

// SendInstallProposal ...
/**
* Sends an install proposal to one or more endorsing peers.
* @param {string} chaincodeName: required - The name of the chaincode.
* @param {[]string} chaincodePath: required - string of the path to the location of the source code of the chaincode
* @param {[]string} chaincodeVersion: required - string of the version of the chaincode
* @param {[]string} chaincodeVersion: optional - Array of byte the chaincodePackage
 */
func (c *chain) SendInstallProposal(chaincodeName string, chaincodePath string, chaincodeVersion string, chaincodePackage []byte) ([]*TransactionProposalResponse, string, error) {

	if chaincodeName == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeName' parameter")
	}
	if chaincodePath == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodePath' parameter")
	}
	if chaincodeVersion == "" {
		return nil, "", fmt.Errorf("Missing 'chaincodeVersion' parameter")
	}

	if chaincodePackage == nil {
		var err error
		chaincodePackage, err = PackageCC(chaincodePath, "")
		if err != nil {
			return nil, "", fmt.Errorf("PackageCC return error: %s", err)
		}
	}

	now := time.Now()
	cds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion}},
		CodePackage: chaincodePackage, EffectiveDate: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}}

	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, "", fmt.Errorf("GetUserContext return error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, "", err
	}

	// create an install from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateInstallProposalFromCDS(cds, creatorID)
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
	}, 0)
	return transactionProposalResponse, txID, err
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
	args []string, chaincodePath string, chaincodeVersion string) ([]*TransactionProposalResponse, string, error) {

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

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, "", fmt.Errorf("GetUserContext return error: %s", err)
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
	}, 0)

	return transactionProposalResponse, txID, err
}

func (c *chain) signPayload(payload []byte) (*SignedEnvelope, error) {
	//Get user info
	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, fmt.Errorf("GetUserContext returned error: %s", err)
	}

	signature, err := c.signObjectWithKey(payload, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil)
	if err != nil {
		return nil, err
	}
	// here's the envelope
	return &SignedEnvelope{Payload: payload, signature: signature}, nil
}

//broadcastEnvelope will send the given envelope to each orderer
func (c *chain) broadcastEnvelope(envelope *SignedEnvelope) ([]*TransactionResponse, error) {
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

// signObjectWithKey will sign the given object with the given key,
// hashOpts and signerOpts
func (c *chain) signObjectWithKey(object []byte, key bccsp.Key,
	hashOpts bccsp.HashOpts, signerOpts bccsp.SignerOpts) ([]byte, error) {
	cryptoSuite := c.clientContext.GetCryptoSuite()
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
	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, fmt.Errorf("GetUserContext return error: %s", err)
	}

	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, err
	}

	signature, err := c.signObjectWithKey(proposalBytes, user.GetPrivateKey(), &bccsp.SHAOpts{}, nil)
	if err != nil {
		return nil, err
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
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
