/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// ChannelClientOpts provides options for creating channel client
// This struct is deprecated.
type ChannelClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ChannelMgmtClientOpts provides options for creating channel management client
// This struct is deprecated.
type ChannelMgmtClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ResourceMgmtClientOpts provides options for creating resource management client
// This struct is deprecated.
type ResourceMgmtClientOpts struct {
	OrgName        string
	TargetFilter   resmgmt.TargetFilter
	ConfigProvider apiconfig.Config
}

// NewSystemClient returns a new client for the system (operations not on a channel)
// TODO: Reduced immutable interface
// TODO: Parameter for setting up the peers
// This function is deprecated.
func (sdk *FabricSDK) NewSystemClient(s apisdk.Session) (apifabclient.FabricClient, error) {
	return sdk.FabricProvider().NewClient(s.Identity())
}

// NewChannelMgmtClientWithOpts returns a new client for managing channels with options
// This function is deprecated.
func (sdk *FabricSDK) NewChannelMgmtClientWithOpts(userName string, opt *ChannelMgmtClientOpts) (chmgmt.ChannelMgmtClient, error) {
	o := []ClientOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}

	c, err := sdk.NewClient(WithUser(userName), o...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.ChannelMgmt()
}

// NewResourceMgmtClientWithOpts returns a new resource management client (user has to be pre-enrolled)
// This function is deprecated.
func (sdk *FabricSDK) NewResourceMgmtClientWithOpts(userName string, opt *ResourceMgmtClientOpts) (resmgmt.ResourceMgmtClient, error) {
	o := []ClientOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.TargetFilter != nil {
		o = append(o, WithTargetFilter(opt.TargetFilter))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}

	c, err := sdk.NewClient(WithUser(userName), o...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.ResourceMgmt()
}

// NewChannelClientWithOpts returns a new client for a channel (user has to be pre-enrolled)
// This function is deprecated.
func (sdk *FabricSDK) NewChannelClientWithOpts(channelID string, userName string, opt *ChannelClientOpts) (apitxn.ChannelClient, error) {
	o := []ClientOption{}
	if opt.OrgName != "" {
		o = append(o, WithOrg(opt.OrgName))
	}
	if opt.ConfigProvider != nil {
		o = append(o, withConfig(opt.ConfigProvider))
	}

	c, err := sdk.NewClient(WithUser(userName), o...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.Channel(channelID)
}

// NewPreEnrolledUserSession returns a new pre-enrolled user session
// This function is deprecated.
func (sdk *FabricSDK) NewPreEnrolledUserSession(orgID string, id string) (*Session, error) {
	return sdk.newSessionFromIdentityName(orgID, id)
}
