/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mspCfg "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
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
}

// ChannelCfg contains channel configuration
type ChannelCfg interface {
	ID() string
	MSPs() []*mspCfg.MSPConfig
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
