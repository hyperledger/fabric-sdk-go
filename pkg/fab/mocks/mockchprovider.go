/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockChannelProvider holds a mock channel provider.
type MockChannelProvider struct {
	ctx                    core.Providers
	transactor             fab.Transactor
	customSelectionService fab.ChannelService
}

// MockChannelService holds a mock channel service.
type MockChannelService struct {
	provider     *MockChannelProvider
	channelID    string
	transactor   fab.Transactor
	mockOrderers []string
}

// NewMockChannelProvider returns a mock ChannelProvider
func NewMockChannelProvider(ctx context.Client) (*MockChannelProvider, error) {
	// Create a mock client with the mock channel
	cp := MockChannelProvider{
		ctx: ctx,
	}
	return &cp, nil
}

// SetTransactor sets the default transactor for all mock channel services
func (cp *MockChannelProvider) SetTransactor(transactor fab.Transactor) {
	cp.transactor = transactor
}

// ChannelService returns a mock ChannelService
func (cp *MockChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {

	if cp.customSelectionService != nil {
		return cp.customSelectionService, nil
	}

	cs := MockChannelService{
		provider:   cp,
		channelID:  channelID,
		transactor: cp.transactor,
	}
	return &cs, nil
}

// SetCustomChannelService sets custom channel service for unit-test purposes
func (cp *MockChannelProvider) SetCustomChannelService(customSelectionService fab.ChannelService) {
	cp.customSelectionService = customSelectionService
}

// SetOrderers sets orderes to mock channel service for unit-test purposes
func (cs *MockChannelService) SetOrderers(orderers []string) {
	cs.mockOrderers = orderers
}

// EventService returns a mock event service
func (cs *MockChannelService) EventService() (fab.EventService, error) {
	return NewMockEventService(), nil
}

// SetTransactor changes the return value of Transactor
func (cs *MockChannelService) SetTransactor(t fab.Transactor) {
	cs.transactor = t
}

// Config ...
func (cs *MockChannelService) Config() (fab.ChannelConfig, error) {
	return nil, nil
}

// Membership returns member identification
func (cs *MockChannelService) Membership() (fab.ChannelMembership, error) {
	return NewMockMembership(), nil
}

//ChannelConfig returns channel config
func (cs *MockChannelService) ChannelConfig() (fab.ChannelCfg, error) {
	return &MockChannelCfg{MockID: cs.channelID, MockOrderers: cs.mockOrderers}, nil
}
