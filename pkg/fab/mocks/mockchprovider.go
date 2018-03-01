/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// MockChannelProvider holds a mock channel provider.
type MockChannelProvider struct {
	ctx        context.ProviderContext
	transactor fab.Transactor
}

// MockChannelService holds a mock channel service.
type MockChannelService struct {
	provider   *MockChannelProvider
	channelID  string
	transactor fab.Transactor
}

// NewMockChannelProvider returns a mock ChannelProvider
func NewMockChannelProvider(ctx context.Context) (*MockChannelProvider, error) {
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
func (cp *MockChannelProvider) ChannelService(ic context.IdentityContext, channelID string) (fab.ChannelService, error) {
	cs := MockChannelService{
		provider:   cp,
		channelID:  channelID,
		transactor: cp.transactor,
	}
	return &cs, nil
}

// EventHub ...
func (cs *MockChannelService) EventHub() (fab.EventHub, error) {
	return NewMockEventHub(), nil
}

// Transactor ...
func (cs *MockChannelService) Transactor() (fab.Transactor, error) {
	return cs.transactor, nil
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

// Ledger ...
func (cs *MockChannelService) Ledger() (fab.ChannelLedger, error) {
	return nil, nil
}
