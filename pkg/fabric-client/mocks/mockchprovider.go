/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/pkg/errors"
)

// MockChannelProvider holds a mock channel provider.
type MockChannelProvider struct {
	ctx        fab.ProviderContext
	channels   map[string]fab.Channel
	transactor fab.Transactor
}

// MockChannelService holds a mock channel service.
type MockChannelService struct {
	provider   *MockChannelProvider
	channelID  string
	transactor fab.Transactor
}

// NewMockChannelProvider returns a mock ChannelProvider
func NewMockChannelProvider(ctx fab.Context) (*MockChannelProvider, error) {
	channels := make(map[string]fab.Channel)

	// Create a mock client with the mock channel
	cp := MockChannelProvider{
		ctx:      ctx,
		channels: channels,
	}
	return &cp, nil
}

// SetChannel convenience method to set channel
func (cp *MockChannelProvider) SetChannel(id string, channel fab.Channel) {
	cp.channels[id] = channel
}

// SetTransactor sets the default transactor for all mock channel services
func (cp *MockChannelProvider) SetTransactor(transactor fab.Transactor) {
	cp.transactor = transactor
}

// NewChannelService returns a mock ChannelService
func (cp *MockChannelProvider) NewChannelService(ic fab.IdentityContext, channelID string) (fab.ChannelService, error) {
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

// Channel ...
func (cs *MockChannelService) Channel() (fab.Channel, error) {
	ch, ok := cs.provider.channels[cs.channelID]
	if !ok {
		return nil, errors.New("No channel")
	}

	return ch, nil
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

// Ledger ...
func (cs *MockChannelService) Ledger() (fab.ChannelLedger, error) {
	return nil, nil
}
