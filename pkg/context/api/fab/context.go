/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// IdentityContext supplies the serialized identity and key reference.
//
// TODO - refactor SigningIdentity and this interface.
type IdentityContext interface {
	MspID() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
}

// ChannelService supplies services related to a channel.
type ChannelService interface {
	Config() (ChannelConfig, error)
	Transactor() (Transactor, error)
	EventService() (EventService, error)
	Membership() (ChannelMembership, error)
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
