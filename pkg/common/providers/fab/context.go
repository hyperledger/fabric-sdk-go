/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

// ChannelService supplies services related to a channel.
type ChannelService interface {
	Config() (ChannelConfig, error)
	EventService(opts ...options.Opt) (EventService, error)
	Membership() (ChannelMembership, error)
	ChannelConfig() (ChannelCfg, error)
	Transactor(reqCtx reqContext.Context) (Transactor, error)
	Discovery() (DiscoveryService, error)
	Selection() (SelectionService, error)
}

// Transactor supplies methods for sending transaction proposals and transactions.
type Transactor interface {
	Sender
	ProposalSender
}

// ChannelProvider supplies Channel related-objects for the named channel.
type ChannelProvider interface {
	ChannelService(ctx ClientContext, channelID string) (ChannelService, error)
}

// ErrorHandler is invoked when an error occurs in one of the services
type ErrorHandler func(ctxt ClientContext, channelID string, err error)
