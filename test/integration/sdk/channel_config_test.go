/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
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
	defer sdk.Close()

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

	confProvider := config.FromFile(testSetup.ConfigFile)
	// Create SDK setup for channel client with retrieve channel configuration from orderer
	sdk, err := fabsdk.New(confProvider,
		fabsdk.WithCorePkg(&ChannelConfigFromOrdererProviderFactory{orderer: setupOrderer(t, confProvider, "orderer.example.com:7050")}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

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
	orderer fab.Orderer
}

// CustomFabricProvider overrides channel config default implementation
type CustomFabricProvider struct {
	*fabpvdr.FabricProvider
	orderer         fab.Orderer
	providerContext context.Providers
}

// CreateChannelConfig initializes the channel config
func (f *CustomFabricProvider) CreateChannelConfig(ic fab.IdentityContext, channelID string) (fab.ChannelConfig, error) {
	ctx := chconfig.Context{
		Providers: f.providerContext,
		Identity:  ic,
	}

	return chconfig.New(ctx, channelID, chconfig.WithOrderer(f.orderer))
}

// CreateFabricProvider returns a new default implementation of fabric primitives
func (f *ChannelConfigFromOrdererProviderFactory) CreateFabricProvider(context context.Providers) (fab.InfraProvider, error) {

	fabProvider := fabpvdr.New(context)

	cfp := CustomFabricProvider{
		FabricProvider:  fabProvider,
		providerContext: context,
		orderer:         f.orderer,
	}
	return &cfp, nil
}

func setupOrderer(t *testing.T, confProvider core.ConfigProvider, address string) fab.Orderer {
	conf, err := confProvider()
	assert.Nil(t, err)

	//Get orderer config by orderer address
	oCfg, err := conf.OrdererConfig(resolveOrdererAddress(address))
	assert.Nil(t, err)

	o, err := orderer.New(conf, orderer.FromOrdererConfig(oCfg))
	assert.Nil(t, err)

	return o
}

// resolveOrdererAddress resolves order address to remove port from address if present
func resolveOrdererAddress(ordererAddress string) string {
	s := strings.Split(ordererAddress, ":")
	if len(s) > 1 {
		return s[0]
	}
	return ordererAddress
}
