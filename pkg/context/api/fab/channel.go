/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mspCfg "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
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
	QueryTransaction(transactionID TransactionID) (*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error)
	QueryByChaincode(request ChaincodeInvokeRequest) ([][]byte, error)
	QueryBySystemChaincode(request ChaincodeInvokeRequest) ([][]byte, error)
	QueryConfigBlock(targets []ProposalProcessor, minResponses int) (*common.ConfigEnvelope, error)
}

// ChannelLedger provides access to the underlying ledger for a channel.
type ChannelLedger interface {
	QueryInfo(targets []ProposalProcessor) ([]*BlockchainInfoResponse, error)
	QueryBlock(blockNumber int, targets []ProposalProcessor) ([]*common.Block, error)
	QueryBlockByHash(blockHash []byte, targets []ProposalProcessor) ([]*common.Block, error)
	QueryTransaction(transactionID TransactionID, targets []ProposalProcessor) ([]*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes(targets []ProposalProcessor) ([]*pb.ChaincodeQueryResponse, error)
	QueryConfigBlock(targets []ProposalProcessor, minResponses int) (*common.ConfigEnvelope, error) // TODO: generalize minResponses
}

// OrgAnchorPeer contains information about an anchor peer on this channel
type OrgAnchorPeer struct {
	Org  string
	Host string
	Port int32
}

// ChannelConfig allows for interaction with peer regarding channel configuration
type ChannelConfig interface {

	// Query channel configuration
	Query() (ChannelCfg, error)
}

// ChannelCfg contains channel configuration
type ChannelCfg interface {
	Name() string
	Msps() []*mspCfg.MSPConfig
	AnchorPeers() []*OrgAnchorPeer
	Orderers() []string
	Versions() *Versions
}

// ChannelMembership helps identify a channel's members
type ChannelMembership interface {
	// Validate if the given ID was issued by the channel's members
	Validate(serializedID []byte) error
	// Verify the given signature
	Verify(serializedID []byte, msg []byte, sig []byte) error
}

// Versions ...
type Versions struct {
	ReadSet  *common.ConfigGroup
	WriteSet *common.ConfigGroup
	Channel  *common.ConfigGroup
}

// BlockchainInfoResponse wraps blockchain info with endorser info
type BlockchainInfoResponse struct {
	BCI      *common.BlockchainInfo
	Endorser string
	Status   int32
}
