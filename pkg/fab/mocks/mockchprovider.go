/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockChannelProvider holds a mock channel provider.
type MockChannelProvider struct {
	ctx                  core.Providers
	customChannelService fab.ChannelService
}

// MockChannelService holds a mock channel service.
type MockChannelService struct {
	provider     *MockChannelProvider
	channelID    string
	transactor   fab.Transactor
	mockOrderers []string
	discovery    fab.DiscoveryService
	selection    fab.SelectionService
	membership   fab.ChannelMembership
}

// NewMockChannelProvider returns a mock ChannelProvider
func NewMockChannelProvider(ctx core.Providers) (*MockChannelProvider, error) {
	// Create a mock client with the mock channel
	cp := MockChannelProvider{
		ctx: ctx,
	}
	return &cp, nil
}

// ChannelService returns a mock ChannelService
func (cp *MockChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {

	if cp.customChannelService != nil {
		return cp.customChannelService, nil
	}

	cs := MockChannelService{
		provider:   cp,
		channelID:  channelID,
		transactor: &MockTransactor{},
		discovery:  NewMockDiscoveryService(nil),
		selection:  NewMockSelectionService(nil),
	}
	return &cs, nil
}

// SetCustomChannelService sets custom channel service for unit-test purposes
func (cp *MockChannelProvider) SetCustomChannelService(customChannelService fab.ChannelService) {
	cp.customChannelService = customChannelService
}

// SetOrderers sets orderes to mock channel service for unit-test purposes
func (cs *MockChannelService) SetOrderers(orderers []string) {
	cs.mockOrderers = orderers
}

// EventService returns a mock event service
func (cs *MockChannelService) EventService(opts ...options.Opt) (fab.EventService, error) {
	return NewMockEventService(), nil
}

// SetTransactor changes the return value of Transactor
func (cs *MockChannelService) SetTransactor(t fab.Transactor) {
	cs.transactor = t
}

// Transactor returns a mock transactor
func (cs *MockChannelService) Transactor(reqCtx reqContext.Context) (fab.Transactor, error) {
	if cs.transactor != nil {
		return cs.transactor, nil
	}
	return &MockTransactor{ChannelID: cs.channelID, Ctx: reqCtx}, nil
}

// Config ...
func (cs *MockChannelService) Config() (fab.ChannelConfig, error) {
	return nil, nil
}

// Membership returns member identification
func (cs *MockChannelService) Membership() (fab.ChannelMembership, error) {
	if cs.membership != nil {
		return cs.membership, nil
	}
	return NewMockMembership(), nil
}

//SetCustomMembership sets custom channel membership for unit-test purposes
func (cs *MockChannelService) SetCustomMembership(customMembership fab.ChannelMembership) {
	cs.membership = customMembership
}

//ChannelConfig returns channel config
func (cs *MockChannelService) ChannelConfig() (fab.ChannelCfg, error) {
	return &MockChannelCfg{MockID: cs.channelID, MockOrderers: cs.mockOrderers}, nil
}

// Discovery returns a mock DiscoveryService
func (cs *MockChannelService) Discovery() (fab.DiscoveryService, error) {
	return cs.discovery, nil
}

// SetDiscovery sets the mock DiscoveryService
func (cs *MockChannelService) SetDiscovery(discovery fab.DiscoveryService) {
	cs.discovery = discovery
}

// Selection returns a mock SelectionService
func (cs *MockChannelService) Selection() (fab.SelectionService, error) {
	return cs.selection, nil
}

// SetSelection sets the mock SelectionService
func (cs *MockChannelService) SetSelection(selection fab.SelectionService) {
	cs.selection = selection
}
