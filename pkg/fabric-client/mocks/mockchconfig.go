/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

// MockChannelCfg contains mock channel configuration
type MockChannelCfg struct {
	MockName        string
	MockMsps        []*msp.MSPConfig
	MockAnchorPeers []*fab.OrgAnchorPeer
	MockOrderers    []string
	MockVersions    *fab.Versions
}

// NewMockChannelCfg ...
func NewMockChannelCfg(name string) fab.ChannelCfg {
	return &MockChannelCfg{MockName: name}
}

// Name returns name
func (cfg *MockChannelCfg) Name() string {
	return cfg.MockName
}

// Msps returns msps
func (cfg *MockChannelCfg) Msps() []*msp.MSPConfig {
	return cfg.MockMsps
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

// MockChannelConfig mocks query channel configuration
type MockChannelConfig struct {
	channelID string
	ctx       fab.Context
}

// NewMockChannelConfig mocks channel config implementation
func NewMockChannelConfig(ctx fab.Context, channelID string) (*MockChannelConfig, error) {
	return &MockChannelConfig{channelID: channelID, ctx: ctx}, nil
}

// Query mocks query for channel configuration
func (c *MockChannelConfig) Query() (fab.ChannelCfg, error) {
	return NewMockChannelCfg(c.channelID), nil
}
