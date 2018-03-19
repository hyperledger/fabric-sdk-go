/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

// MockChannelCfg contains mock channel configuration
type MockChannelCfg struct {
	MockID          string
	MockMSPs        []*msp.MSPConfig
	MockAnchorPeers []*fab.OrgAnchorPeer
	MockOrderers    []string
	MockVersions    *fab.Versions
	MockMembership  fab.ChannelMembership
}

// NewMockChannelCfg ...
func NewMockChannelCfg(id string) *MockChannelCfg {
	return &MockChannelCfg{MockID: id}
}

// ID returns name
func (cfg *MockChannelCfg) ID() string {
	return cfg.MockID
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
