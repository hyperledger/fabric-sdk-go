// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	mockCore "github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
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
	sdk.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

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

	sdk1.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk1.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	//prepare config backend for sdk2

	customBackend, err := getCustomBackend()
	if err != nil {
		t.Fatalf("failed to get configbackend for test: %s", err)
	}
	configProvider := func() ([]core.ConfigBackend, error) {
		return customBackend, nil
	}

	sdk2, err := New(configProvider)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk2.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Default sdk with two channels
	configBackend, err := sdk1.Config()
	if err != nil {
		t.Fatalf("Error getting config backend from sdk: %s", err)
	}
	checkClientOrg(configBackend, t, sdkValidClientOrg1)

	configBackend, err = sdk2.Config()
	if err != nil {
		t.Fatalf("Error getting config backend from sdk: %s", err)
	}
	checkClientOrg(configBackend, t, sdkValidClientOrg2)

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

func checkClientOrg(configBackend core.ConfigBackend, t *testing.T, orgName string) {
	identityConfig, err := msp.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("Error getting identity config : %s", err)
	}
	client := identityConfig.Client()
	if client.Organization != orgName {
		t.Fatalf("Unexpected org in config: %s", client.Organization)
	}
}

func getCustomBackend() ([]core.ConfigBackend, error) {
	backend, err := config.FromFile(sdkConfigFile)()
	if err != nil {
		return nil, err
	}

	//read existing client config from config
	configLookup := lookup.New(backend...)
	res, ok := configLookup.Lookup("client")
	if !ok {
		return nil, errors.New("failed to created custom backend for test")
	}
	resMap := res.(map[string]interface{})
	//update it
	resMap["organization"] = "org2"

	//set it to backend map
	backendMap := make(map[string]interface{})
	backendMap["client"] = resMap

	backends := append([]core.ConfigBackend{}, &mockCore.MockConfigBackend{KeyValueMap: backendMap})
	return append(backends, backend...), nil
}

// ClientConfig provides the definition of the client configuration
type customClientConfig struct {
	Organization string
	TLSCerts     clientTLSConfig
}

type clientTLSConfig struct {
	//Client TLS information
	Client endpoint.TLSKeyPair
}
