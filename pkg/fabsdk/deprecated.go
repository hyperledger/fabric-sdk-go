/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/resmgmtclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/pkg/errors"
)

// ChannelClientOpts provides options for creating channel client
//
// Deprecated: Use NewClient instead.
type ChannelClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ChannelMgmtClientOpts provides options for creating channel management client
//
// Deprecated: Use NewClient instead.
type ChannelMgmtClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ResourceMgmtClientOpts provides options for creating resource management client
//
// Deprecated: Use NewClient instead.
type ResourceMgmtClientOpts struct {
	OrgName        string
	TargetFilter   resmgmt.TargetFilter
	ConfigProvider apiconfig.Config
}

// NewChannelClientWithOpts returns a new client for a channel (user has to be pre-enrolled)
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewChannelClientWithOpts(channelID string, userName string, opt *ChannelClientOpts) (chclient.ChannelClient, error) {
	o := []ContextOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}

	c := sdk.NewClient(WithUser(userName), o...)
	return c.Channel(channelID)
}

// NewChannelClient returns a new client for a channel
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewChannelClient(channelID string, userName string, opts ...ContextOption) (chclient.ChannelClient, error) {
	c := sdk.NewClient(WithUser(userName), opts...)
	return c.Channel(channelID)
}

// NewPreEnrolledUser returns a new pre-enrolled user
func (sdk *FabricSDK) NewPreEnrolledUser(orgID string, userName string) (apifabclient.IdentityContext, error) {
	return sdk.newUser(orgID, userName)
}

// newSessionFromIdentityName returns a new user session
func (sdk *FabricSDK) newSessionFromIdentityName(orgID string, id string) (*session, error) {

	user, err := sdk.newUser(orgID, id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user")
	}

	session := newSession(user, sdk.channelProvider)

	return session, nil
}

// NewSystemClient returns a new client for the system (operations not on a channel)
//
// Deprecated: the system client is being replaced with the interfaces supplied by NewClient()
func (sdk *FabricSDK) NewSystemClient(s apisdk.SessionContext) (apifabclient.Resource, error) {
	return sdk.fabricProvider.CreateResourceClient(s)
}
