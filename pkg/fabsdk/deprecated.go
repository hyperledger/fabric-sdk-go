/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
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

// NewChannelMgmtClientWithOpts returns a new client for managing channels with options
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewChannelMgmtClientWithOpts(userName string, opt *ChannelMgmtClientOpts) (chmgmt.ChannelMgmtClient, error) {
	o := []ContextOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}

	c := sdk.NewClient(WithUser(userName), o...)
	return c.ChannelMgmt()
}

// NewResourceMgmtClientWithOpts returns a new resource management client (user has to be pre-enrolled)
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewResourceMgmtClientWithOpts(userName string, opt *ResourceMgmtClientOpts) (resmgmt.ResourceMgmtClient, error) {
	o := []ContextOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}
	clientOpts := []ClientOption{}
	if opt.TargetFilter != nil {
		clientOpts = append(clientOpts, WithTargetFilter(opt.TargetFilter))
	}
	c := sdk.NewClient(WithUser(userName), o...)
	return c.ResourceMgmt(clientOpts...)
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

// NewChannelMgmtClient returns a new client for managing channels
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewChannelMgmtClient(userName string, opts ...ContextOption) (chmgmt.ChannelMgmtClient, error) {
	c := sdk.NewClient(WithUser(userName), opts...)
	return c.ChannelMgmt()
}

// NewResourceMgmtClient returns a new client for managing system resources
//
// Deprecated: Use NewClient instead.
func (sdk *FabricSDK) NewResourceMgmtClient(userName string, opts ...ContextOption) (resmgmt.ResourceMgmtClient, error) {
	c := sdk.NewClient(WithUser(userName), opts...)
	return c.ResourceMgmt()
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
