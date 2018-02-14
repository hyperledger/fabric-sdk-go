/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() SigningManager
	Config() config.Config
	CryptoSuite() apicryptosuite.CryptoSuite
}

// ChannelProvider supplies Channel related-objects for the named channel.
type ChannelProvider interface {
	NewChannelService(ic IdentityContext, channelID string) (ChannelService, error)
}

// ChannelService supplies services related to a channel.
type ChannelService interface {
	Config() (ChannelConfig, error)
	Ledger() (ChannelLedger, error)
	Channel() (Channel, error)
	EventHub() (EventHub, error) // TODO support new event delivery
}
