/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockInfraProvider represents the default implementation of Fabric objects.
type MockInfraProvider struct {
	customOrderer fab.Orderer
}

// CreateEventService creates the event service.
func (f *MockInfraProvider) CreateEventService(ic fab.ClientContext, channelID string, opts ...options.Opt) (fab.EventService, error) {
	panic("not implemented")
}

// CreateChannelCfg creates the channel configuration
func (f *MockInfraProvider) CreateChannelCfg(ctx fab.ClientContext, name string) (fab.ChannelCfg, error) {
	return nil, nil
}

// CreateChannelMembership returns a channel member identifier
func (f *MockInfraProvider) CreateChannelMembership(ctx fab.ClientContext, channel string) (fab.ChannelMembership, error) {
	return nil, fmt.Errorf("Not implemented")
}

// CreateChannelConfig initializes the channel config
func (f *MockInfraProvider) CreateChannelConfig(channelID string) (fab.ChannelConfig, error) {
	return nil, nil
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *MockInfraProvider) CreatePeerFromConfig(peerCfg *fab.NetworkPeer) (fab.Peer, error) {
	if peerCfg != nil {
		p := NewMockPeer(peerCfg.MSPID, peerCfg.URL)
		p.SetMSPID(peerCfg.MSPID)
		p.SetProperties(peerCfg.Properties)

		return p, nil
	}
	return &MockPeer{}, nil
}

// CreateOrdererFromConfig creates a default implementation of Orderer based on configuration.
func (f *MockInfraProvider) CreateOrdererFromConfig(cfg *fab.OrdererConfig) (fab.Orderer, error) {
	if f.customOrderer != nil {
		return f.customOrderer, nil
	}

	if cfg.URL != "" {
		return &MockOrderer{OrdererURL: cfg.URL}, nil
	}

	return &MockOrderer{}, nil
}

//CommManager returns comm provider
func (f *MockInfraProvider) CommManager() fab.CommManager {
	return nil
}

// SetCustomOrderer creates a default implementation of Orderer based on configuration.
func (f *MockInfraProvider) SetCustomOrderer(customOrderer fab.Orderer) {
	f.customOrderer = customOrderer
}

//Close mock close function
func (f *MockInfraProvider) Close() {
}
