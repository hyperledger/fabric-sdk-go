/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
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
	Sender
	ProposalSender

	Name() string
	ChannelConfig() (*common.ConfigEnvelope, error)

	SendInstantiateProposal(chaincodeName string, args [][]byte, chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope,
		collConfig []*common.CollectionConfig, targets []ProposalProcessor) ([]*TransactionProposalResponse, TransactionID, error)
	SendUpgradeProposal(chaincodeName string, args [][]byte, chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope, targets []ProposalProcessor) ([]*TransactionProposalResponse, TransactionID, error)

	// Network
	// Deprecated: getters/setters are deprecated from interface.
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
	OrganizationUnits() ([]string, error)

	// Query
	QueryInfo() (*common.BlockchainInfo, error)
	QueryBlock(blockNumber int) (*common.Block, error)
	QueryBlockByHash(blockHash []byte) (*common.Block, error)
	QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error)
	QueryByChaincode(ChaincodeInvokeRequest) ([][]byte, error)
	QueryBySystemChaincode(request ChaincodeInvokeRequest) ([][]byte, error)
	QueryConfigBlock(peers []Peer, minResponses int) (*common.ConfigEnvelope, error)
}

// ChannelLedger provides access to the underlying ledger for a channel.
type ChannelLedger interface {
	QueryInfo(targets []ProposalProcessor) ([]*common.BlockchainInfo, error)
	QueryBlock(blockNumber int, targets []ProposalProcessor) ([]*common.Block, error)
	QueryBlockByHash(blockHash []byte, targets []ProposalProcessor) ([]*common.Block, error)
	QueryTransaction(transactionID string, targets []ProposalProcessor) ([]*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes(targets []ProposalProcessor) ([]*pb.ChaincodeQueryResponse, error)
}

// OrgAnchorPeer contains information about an anchor peer on this channel
type OrgAnchorPeer struct {
	Org  string
	Host string
	Port int32
}
