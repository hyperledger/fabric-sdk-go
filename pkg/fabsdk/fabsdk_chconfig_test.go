// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

const (
	sdkValidClientOrg2 = "org2"
)

func TestNewDefaultSDK(t *testing.T) {
	// Test New SDK with valid config file
	sdk, err := New(config.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func verifySDK(t *testing.T, sdk *FabricSDK) {

	// Mock channel provider cache
	sdk.provider.InfraProvider().(*fabpvdr.InfraProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk.provider.InfraProvider().(*fabpvdr.InfraProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Get a common client context for the following tests
	chCtx := sdk.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg2))

	// Test new channel client with options
	_, err := channel.New(chCtx)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
}

func TestWithConfigOpt(t *testing.T) {
	// Test New SDK with valid config file
	c := config.FromFile(sdkConfigFile)

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func TestNewDefaultTwoValidSDK(t *testing.T) {
	sdk1, err := New(config.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache

	sdk1.provider.InfraProvider().(*fabpvdr.InfraProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk1.provider.InfraProvider().(*fabpvdr.InfraProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	sdk2, err := New(config.FromFile("./testdata/test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk2.provider.InfraProvider().(*fabpvdr.InfraProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Default sdk with two channels
	configBackend, err := sdk1.Config()
	if err != nil {
		t.Fatalf("Error getting config backend from sdk: %s", err)
	}

	identityConfig, err := msp.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("Error getting identity config: %s", err)
	}

	client1, err := identityConfig.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != sdkValidClientOrg1 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	configBackend, err = sdk2.Config()
	if err != nil {
		t.Fatalf("Error getting config backend from sdk: %s", err)
	}

	identityConfig, err = msp.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("Error getting identity config : %s", err)
	}

	client2, err := identityConfig.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != sdkValidClientOrg2 {
		t.Fatalf("Unexpected org in config: %s", client2.Organization)
	}

	// Get a common client context for the following tests
	//cc1 := sdk1.NewClient(WithUser(sdkValidClientUser))

	cc1CtxC1 := sdk1.ChannelContext("mychannel", WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg1))
	cc1CtxC2 := sdk1.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg1))

	// Test SDK1 channel clients ('mychannel', 'orgchannel')
	_, err = channel.New(cc1CtxC1)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = channel.New(cc1CtxC2)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Get a common client context for the following tests
	cc2CtxC := sdk1.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg2))

	// SDK 2 has 'orgchannel' configured
	_, err = channel.New(cc2CtxC)
	if err != nil {
		t.Fatalf("Failed to create new 'orgchannel' channel client: %s", err)
	}
}
