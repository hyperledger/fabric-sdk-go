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
	SigningUser fab.User
}

// SaveChannelOpts contains options for saving channel
type SaveChannelOpts struct {
	OrdererID string // use specific orderer
}

// ChannelMgmtClient supports creating new channels
type ChannelMgmtClient interface {

	// SaveChannel creates or updates channel
	SaveChannel(req SaveChannelRequest) error

	// SaveChannel creates or updates channel with custom options
	SaveChannelWithOpts(req SaveChannelRequest, opts SaveChannelOpts) error
}
