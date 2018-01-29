/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chmgmtclient

import fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

// SaveChannelRequest contains parameters for creating or updating channel
type SaveChannelRequest struct {
	// Channel Name (ID)
	ChannelID string
	// Path to channel configuration file
	ChannelConfig string
	// User that signs channel configuration
	SigningIdentity fab.IdentityContext
}

// Opts contains options for saving channel, this struct is intended for reference only.
// Will be used in form of Option to pass arguments
type Opts struct {
	OrdererID string // use specific orderer
}

//Option func for each Opts argument
type Option func(opts *Opts) error

// ChannelMgmtClient supports creating new channels
type ChannelMgmtClient interface {

	// SaveChannel creates or updates channel with optional Opts options
	SaveChannel(req SaveChannelRequest, options ...Option) error
}
