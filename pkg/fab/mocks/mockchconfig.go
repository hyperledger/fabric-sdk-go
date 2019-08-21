/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContext "context"

	"github.com/hyperledger/fabric-protos-go/common"

	msp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockChannelCfg contains mock channel configuration
type MockChannelCfg struct {
	MockID           string
	MockBlockNumber  uint64
	MockMSPs         []*msp.MSPConfig
	MockAnchorPeers  []*fab.OrgAnchorPeer
	MockOrderers     []string
	MockVersions     *fab.Versions
	MockMembership   fab.ChannelMembership
	MockCapabilities map[fab.ConfigGroupKey]map[string]bool
}

// NewMockChannelCfg ...
func NewMockChannelCfg(id string) *MockChannelCfg {
	capabilities := make(map[fab.ConfigGroupKey]map[string]bool)
	capabilities[fab.ChannelGroupKey] = make(map[string]bool)
	capabilities[fab.ApplicationGroupKey] = make(map[string]bool)
	capabilities[fab.OrdererGroupKey] = make(map[string]bool)

	return &MockChannelCfg{
		MockID:           id,
		MockCapabilities: capabilities,
	}
}

// ID returns name
func (cfg *MockChannelCfg) ID() string {
	return cfg.MockID
}

// BlockNumber returns block number
func (cfg *MockChannelCfg) BlockNumber() uint64 {
	return cfg.MockBlockNumber
}

// MSPs returns msps
func (cfg *MockChannelCfg) MSPs() []*msp.MSPConfig {
	return cfg.MockMSPs
}

// AnchorPeers returns anchor peers
func (cfg *MockChannelCfg) AnchorPeers() []*fab.OrgAnchorPeer {
	return cfg.MockAnchorPeers
}

// Orderers returns orderers
func (cfg *MockChannelCfg) Orderers() []string {
	return cfg.MockOrderers
}

// Versions returns versions
func (cfg *MockChannelCfg) Versions() *fab.Versions {
	return cfg.MockVersions
}

// HasCapability indicates whether or not the given group has the given capability
func (cfg *MockChannelCfg) HasCapability(group fab.ConfigGroupKey, capability string) bool {
	capabilities, ok := cfg.MockCapabilities[group]
	if !ok {
		return false
	}
	return capabilities[capability]
}

// MockChannelConfig mockcore query channel configuration
type MockChannelConfig struct {
	channelID string
	ctx       context.Client
}

// NewMockChannelConfig mockcore channel config implementation
func NewMockChannelConfig(ctx context.Client, channelID string) (*MockChannelConfig, error) {
	return &MockChannelConfig{channelID: channelID, ctx: ctx}, nil
}

// Query mockcore query for channel configuration
func (c *MockChannelConfig) Query(reqCtx reqContext.Context) (fab.ChannelCfg, error) {
	return NewMockChannelCfg(c.channelID), nil
}

// QueryBlock mockcore query for channel configuration block
func (c *MockChannelConfig) QueryBlock(reqCtx reqContext.Context) (*common.Block, error) {
	return &common.Block{}, nil
}
