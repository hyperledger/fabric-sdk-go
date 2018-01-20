/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
)

// Session primarily represents the session and identity context
// TODO: does this need to be exposed?
type Session interface {
	Identity() apifabclient.IdentityContext
}

// Client represents the Client APIs supported by the SDK
type Client interface {
	ChannelMgmt() (chmgmt.ChannelMgmtClient, error)
	ResourceMgmt() (resmgmt.ResourceMgmtClient, error)
	Channel(id string) (apitxn.ChannelClient, error)
}
