/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestChannelConfig(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	cs, err := sdk.NewClient(fabsdk.WithUser("User1")).ChannelService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel service: %s", err)
	}

	cfg, err := cs.Config()
	if err != nil {
		t.Fatalf("Failed to create new channel config: %s", err)
	}

	response, err := cfg.Query()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "orderer.example.com:7050"
	found := false
	for _, o := range response.Orderers() {
		if o == expected {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Expected orderer %s, got %s", expected, response.Orderers())
	}

}

func TestChannelConfigWithOrderer(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	// Create SDK setup for channel client with retrieve channel configuration from orderer
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile),
		fabsdk.WithCorePkg(&ChannelConfigFromOrdererProviderFactory{orderer: "orderer.example.com"}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	cs, err := sdk.NewClient(fabsdk.WithUser("User1")).ChannelService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel service: %s", err)
	}

	cfg, err := cs.Config()
	if err != nil {
		t.Fatalf("Failed to create new channel config: %s", err)
	}

	response, err := cfg.Query()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "orderer.example.com:7050"
	found := false
	for _, o := range response.Orderers() {
		if o == expected {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Expected orderer %s, got %s", expected, response.Orderers())
	}

}

// ChannelConfigFromOrdererProviderFactory is configured to retrieve channel config from orderer
type ChannelConfigFromOrdererProviderFactory struct {
	defcore.ProviderFactory
	orderer string
}

// CustomFabricProvider overrides channel config default implementation
type CustomFabricProvider struct {
	*fabpvdr.FabricProvider
	orderer         string
	providerContext apifabclient.ProviderContext
}

// CreateChannelConfig initializes the channel config
func (f *CustomFabricProvider) CreateChannelConfig(ic apifabclient.IdentityContext, channelID string) (apifabclient.ChannelConfig, error) {
	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}

	return chconfig.New(ctx, channelID, chconfig.WithOrderer(f.orderer))
}

// NewFabricProvider returns a new default implementation of fabric primitives
func (f *ChannelConfigFromOrdererProviderFactory) NewFabricProvider(context apifabclient.ProviderContext) (api.FabricProvider, error) {

	fabProvider := fabpvdr.New(context)

	cfp := CustomFabricProvider{
		FabricProvider:  fabProvider,
		providerContext: context,
		orderer:         f.orderer,
	}
	return &cfp, nil
}
