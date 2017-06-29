/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Channel ...
/**
 * Channel representing a Channel with which the client SDK interacts.
 *
 * The Channel object captures settings for a channel, which is created by
 * the orderers to isolate transactions delivery to peers participating on channel.
 * A channel must be initialized after it has been configured with the list of peers
 * and orderers. The initialization sends a get configuration block request to the
 * primary orderer to retrieve the configuration settings for this channel.
 */
type Channel interface {
	Name() string
	Initialize(data []byte) error
	IsInitialized() bool
	IsSecurityEnabled() bool
	TCertBatchSize() int
	SetTCertBatchSize(batchSize int)
	AddPeer(peer Peer) error
	RemovePeer(peer Peer)
	Peers() []Peer
	AnchorPeers() []OrgAnchorPeer
	SetPrimaryPeer(peer Peer) error
	PrimaryPeer() Peer
	AddOrderer(orderer Orderer) error
	RemoveOrderer(orderer Orderer)
	Orderers() []Orderer
	SetMSPManager(mspManager msp.MSPManager)
	MSPManager() msp.MSPManager
	GenesisBlock(request *GenesisBlockRequest) (*common.Block, error)
	JoinChannel(request *JoinChannelRequest) error
	UpdateChannel() bool
	IsReadonly() bool
	QueryInfo() (*common.BlockchainInfo, error)
	QueryBlock(blockNumber int) (*common.Block, error)
	QueryBlockByHash(blockHash []byte) (*common.Block, error)
	QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error)
	QueryByChaincode(chaincodeName string, args []string, targets []Peer) ([][]byte, error)
	CreateTransactionProposal(chaincodeName string, channelID string, args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error)
	SendTransactionProposal(proposal *TransactionProposal, retry int, targets []Peer) ([]*TransactionProposalResponse, error)
	CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error)
	SendTransaction(tx *Transaction) ([]*TransactionResponse, error)
	SendInstantiateProposal(chaincodeName string, channelID string, args []string, chaincodePath string, chaincodeVersion string, targets []Peer) ([]*TransactionProposalResponse, string, error)
	OrganizationUnits() ([]string, error)
	QueryExtensionInterface() ChannelExtension
	LoadConfigUpdateEnvelope(data []byte) error
}

// The ChannelExtension interface allows extensions of the SDK to add functionality to Channel overloads.
type ChannelExtension interface {
	ClientContext() FabricClient

	SignPayload(payload []byte) (*SignedEnvelope, error)
	BroadcastEnvelope(envelope *SignedEnvelope) ([]*TransactionResponse, error)

	// TODO: This should go somewhere else - see TransactionProposal.GetBytes(). - deprecated
	ProposalBytes(tp *TransactionProposal) ([]byte, error)
}

// OrgAnchorPeer contains information about an anchor peer on this channel
type OrgAnchorPeer struct {
	Org  string
	Host string
	Port int32
}

// GenesisBlockRequest ...
type GenesisBlockRequest struct {
	TxID  string
	Nonce []byte
}

// The TransactionProposal object to be send to the endorsers
type TransactionProposal struct {
	TransactionID string

	SignedProposal *pb.SignedProposal
	Proposal       *pb.Proposal
}

// TransactionProposalResponse ...
/**
 * The TransactionProposalResponse result object returned from endorsers.
 */
type TransactionProposalResponse struct {
	Endorser string
	Err      error
	Status   int32

	Proposal         *TransactionProposal
	ProposalResponse *pb.ProposalResponse
}

// JoinChannelRequest allows a set of peers to transact on a channel on the network
type JoinChannelRequest struct {
	Targets      []Peer
	GenesisBlock *common.Block
	TxID         string
	Nonce        []byte
}

// The Transaction object created from an endorsed proposal
type Transaction struct {
	Proposal    *TransactionProposal
	Transaction *pb.Transaction
}

// TransactionResponse ...
/**
 * The TransactionProposalResponse result object returned from orderers.
 */
type TransactionResponse struct {
	Orderer string
	Err     error
}
