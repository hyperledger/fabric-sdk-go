/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"

	"github.com/hyperledger/fabric-protos-go/common"
	mspCfg "github.com/hyperledger/fabric-protos-go/msp"
)

// OrgAnchorPeer contains information about an anchor peer on this channel
type OrgAnchorPeer struct {
	Org  string
	Host string
	Port int32
}

// ChannelConfig allows for interaction with peer regarding channel configuration
type ChannelConfig interface {

	// Query channel configuration
	Query(reqCtx reqContext.Context) (ChannelCfg, error)

	// QueryBlock queries channel configuration block
	QueryBlock(reqCtx reqContext.Context) (*common.Block, error)
}

// ConfigGroupKey is the config group key
type ConfigGroupKey string

const (
	// ChannelGroupKey is the Channel config group key
	ChannelGroupKey ConfigGroupKey = ""
	// OrdererGroupKey is the Orderer config group key
	OrdererGroupKey ConfigGroupKey = "Orderer"
	// ApplicationGroupKey is the Application config group key
	ApplicationGroupKey ConfigGroupKey = "Application"
)

const (
	// V1_1Capability indicates that Fabric 1.1 features are supported
	V1_1Capability = "V1_1"
	// V1_2Capability indicates that Fabric 1.2 features are supported
	V1_2Capability = "V1_2"
)

// ChannelCfg contains channel configuration
type ChannelCfg interface {
	ID() string
	BlockNumber() uint64
	MSPs() []*mspCfg.MSPConfig
	AnchorPeers() []*OrgAnchorPeer
	Orderers() []string
	Versions() *Versions
	HasCapability(group ConfigGroupKey, capability string) bool
}

// ChannelMembership helps identify a channel's members
type ChannelMembership interface {
	// Validate if the given ID was issued by the channel's members
	Validate(serializedID []byte) error
	// Verify the given signature
	Verify(serializedID []byte, msg []byte, sig []byte) error
	//Check is given MSP is available
	ContainsMSP(msp string) bool
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
