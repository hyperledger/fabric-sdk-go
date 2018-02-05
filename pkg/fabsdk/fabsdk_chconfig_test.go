// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestNewDefaultSDK(t *testing.T) {
	// Test New SDK with valid config file
	sdk, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func verifySDK(t *testing.T, sdk *FabricSDK) {

	// Mock channel provider cache
	sdk.channelProvider.SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk.channelProvider.SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Get a common client context for the following tests
	c := sdk.NewClient(WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg2))

	// Test configuration failure for channel client (mychannel does't have event source configured for Org2)
	_, err := c.Channel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to create channel client since event source not configured for Org2")
	}

	// Test new channel client with options
	_, err = c.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
}

func TestWithConfigOpt(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func TestNewDefaultTwoValidSDK(t *testing.T) {
	sdk1, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk1.channelProvider.SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk1.channelProvider.SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	sdk2, err := New(configImpl.FromFile("./testdata/test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk2.channelProvider.SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Default sdk with two channels
	client1, err := sdk1.config.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != sdkValidClientOrg1 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	client2, err := sdk2.config.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != sdkValidClientOrg2 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	// Get a common client context for the following tests
	cc1 := sdk1.NewClient(WithUser(sdkValidClientUser))

	// Test SDK1 channel clients ('mychannel', 'orgchannel')
	_, err = cc1.Channel("mychannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = cc1.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Get a common client context for the following tests
	cc2 := sdk2.NewClient(WithUser(sdkValidClientUser))

	// SDK 2 doesn't have 'mychannel' configured
	_, err = cc2.Channel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to create channel that is not configured")
	}

	// SDK 2 has 'orgchannel' configured
	_, err = cc2.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new 'orgchannel' channel client: %s", err)
	}
}
