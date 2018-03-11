/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// Client supplies the configuration and signing identity to client objects.
type Client interface {
	Providers
	msp.Identity
}

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	msp.Providers
	fab.Providers
}

//Channel supplies the configuration for channel context client
type Channel interface {
	Client
	DiscoveryService() fab.DiscoveryService
	SelectionService() fab.SelectionService
	ChannelService() fab.ChannelService
}

//ClientProvider returns client context
type ClientProvider func() (Client, error)

//ChannelProvider returns channel client context
type ChannelProvider func() (Channel, error)
