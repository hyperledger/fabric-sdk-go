/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// Client supplies the configuration and signing identity to client objects.
type Client fab.ClientContext

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	msp.Providers
	fab.Providers
}

// Local supplies the configuration for a local context client
type Local interface {
	Client
	LocalDiscoveryService() fab.DiscoveryService
}

// Channel supplies the configuration for channel context client
type Channel interface {
	Client
	ChannelService() fab.ChannelService
	ChannelID() string
}

// ClientProvider returns client context
type ClientProvider func() (Client, error)

// LocalProvider returns local client context
type LocalProvider func() (Local, error)

// ChannelProvider returns channel client context
type ChannelProvider func() (Channel, error)
