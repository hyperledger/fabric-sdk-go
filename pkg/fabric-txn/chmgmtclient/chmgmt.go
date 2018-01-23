/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chmgmtclient enables channel management client
package chmgmtclient

import (
	"io/ioutil"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ChannelMgmtClient enables managing channels in Fabric network.
type ChannelMgmtClient struct {
	provider fab.ProviderContext
	identity fab.IdentityContext
	resource fab.Resource
}

// Context holds the providers and services needed to create a ChannelMgmtClient.
type Context struct {
	fab.ProviderContext
	fab.IdentityContext
	Resource fab.Resource
}

// New returns a channel management client instance
func New(c Context) (*ChannelMgmtClient, error) {
	cc := &ChannelMgmtClient{
		provider: c.ProviderContext,
		identity: c.IdentityContext,
		resource: c.Resource,
	}
	return cc, nil
}

// SaveChannel creates or updates channel
func (cc *ChannelMgmtClient) SaveChannel(req chmgmt.SaveChannelRequest) error {
	return cc.SaveChannelWithOpts(req, chmgmt.SaveChannelOpts{})
}

// SaveChannelWithOpts creates or updates channel with custom options
func (cc *ChannelMgmtClient) SaveChannelWithOpts(req chmgmt.SaveChannelRequest, opts chmgmt.SaveChannelOpts) error {

	if req.ChannelID == "" || req.ChannelConfig == "" {
		return errors.New("must provide channel ID and channel config")
	}

	logger.Debugf("***** Saving channel: %s *****\n", req.ChannelID)

	// Signing user has to belong to one of configured channel organisations
	// In case that order org is one of channel orgs we can use context user
	signer := cc.identity
	if req.SigningIdentity != nil {
		// Retrieve custom signing identity here
		signer = req.SigningIdentity
	}

	if signer == nil {
		return errors.New("must provide signing user")
	}

	configTx, err := ioutil.ReadFile(req.ChannelConfig)
	if err != nil {
		return errors.WithMessage(err, "reading channel config file failed")
	}

	chConfig, err := cc.resource.ExtractChannelConfig(configTx)
	if err != nil {
		return errors.WithMessage(err, "extracting channel config failed")
	}

	configSignature, err := cc.resource.SignChannelConfig(chConfig, signer)
	if err != nil {
		return errors.WithMessage(err, "signing configuration failed")
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	// Figure out orderer configuration
	var ordererCfg *config.OrdererConfig
	if opts.OrdererID != "" {
		ordererCfg, err = cc.provider.Config().OrdererConfig(opts.OrdererID)
	} else {
		// Default is random orderer from configuration
		ordererCfg, err = cc.provider.Config().RandomOrdererConfig()
	}

	// Check if retrieving orderer configuration went ok
	if err != nil || ordererCfg == nil {
		return errors.Errorf("failed to retrieve orderer config: %s", err)
	}

	orderer, err := orderer.New(cc.provider.Config(), orderer.FromOrdererConfig(ordererCfg))
	if err != nil {
		return errors.WithMessage(err, "failed to create new orderer from config")
	}

	request := fab.CreateChannelRequest{
		Name:       req.ChannelID,
		Orderer:    orderer,
		Config:     chConfig,
		Signatures: configSignatures,
	}

	_, err = cc.resource.CreateChannel(request)
	if err != nil {
		return errors.WithMessage(err, "create channel failed")
	}

	return nil
}
