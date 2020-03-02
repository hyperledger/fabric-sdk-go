/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package provider

import (
	"strings"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"

	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

func TestChannelConfig(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	channelCtx, err := org1ChannelClientContext()
	if err != nil {
		t.Fatalf("Failed to get channel client context: %s", err)
	}

	cs := channelCtx.ChannelService()

	cfg, err := cs.Config()
	if err != nil {
		t.Fatalf("Failed to create new channel config: %s", err)
	}

	reqCtx, cancel := context.NewRequest(channelCtx, context.WithTimeoutType(fab.PeerResponse))
	defer cancel()

	block, err := cfg.QueryBlock(reqCtx)
	if err != nil {
		t.Fatal(err)
	}

	checkConfigBlock(t, block)

	response, err := cfg.Query(reqCtx)
	if err != nil {
		t.Fatal(err)
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

func checkConfigBlock(t *testing.T, block *common.Block) {
	if block.Header == nil {
		t.Fatal("expected header in block")
	}

	_, err := resource.CreateConfigEnvelope(block.Data.Data[0])
	if err != nil {
		t.Fatal("expected envelope in block")
	}
}

func TestChannelConfigWithOrderer(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ChannelID:           "mychannel",
		OrgID:               org1Name,
		ChannelConfigTxFile: integration.GetChannelConfigTxPath("mychannel.tx"),
	}

	configBackend, err := integration.ConfigBackend()
	if err != nil {
		t.Fatalf("Unexpected error from config backend: %s", err)
	}

	cryptoSuiteConfig := cryptosuite.ConfigFromBackend(configBackend...)

	endpointConfig, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Unexpected error from config: %s", err)
	}

	identityConfig, err := msp.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Unexpected error from config: %s", err)
	}

	// Create SDK setup for channel client with retrieve channel configuration from orderer
	sdk, err := fabsdk.New(nil, fabsdk.WithCryptoSuiteConfig(cryptoSuiteConfig), fabsdk.WithEndpointConfig(endpointConfig), fabsdk.WithIdentityConfig(identityConfig),
		fabsdk.WithCorePkg(&ChannelConfigFromOrdererProviderFactory{orderer: setupOrderer(t, endpointConfig, "orderer.example.com:7050")}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	if err = testSetup.Initialize(sdk); err != nil {
		t.Fatal(err)
	}

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	channelCtx, err := org1ChannelClientContext()
	if err != nil {
		t.Fatalf("Failed to get channel client context: %s", err)
	}

	cs := channelCtx.ChannelService()

	cfg, err := cs.Config()
	if err != nil {
		t.Fatalf("Failed to create new channel config: %s", err)
	}

	queryChannelCfg(channelCtx, cfg, t)

}

func queryChannelCfg(channelCtx contextAPI.Channel, cfg fab.ChannelConfig, t *testing.T) {
	reqCtx, cancel := context.NewRequest(channelCtx, context.WithTimeoutType(fab.OrdererResponse))
	defer cancel()

	block, err := cfg.QueryBlock(reqCtx)
	if err != nil {
		t.Fatal(err)
	}
	checkConfigBlock(t, block)

	response, err := cfg.Query(reqCtx)
	if err != nil {
		t.Fatal(err)
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

// CustomInfraProvider overrides channel config default implementation
type CustomInfraProvider struct {
	*fabpvdr.InfraProvider
	orderer         fab.Orderer
	providerContext api.Providers
}

// Initialize sets the provider context
func (f *CustomInfraProvider) Initialize(providers contextAPI.Providers) error {
	f.providerContext = providers
	f.InfraProvider.Initialize(providers)
	return nil
}

// CreateChannelConfig initializes the channel config
func (f *CustomInfraProvider) CreateChannelConfig(channelID string) (fab.ChannelConfig, error) {
	return chconfig.New(channelID, chconfig.WithOrderer(f.orderer))
}

// CreateInfraProvider returns a new default implementation of fabric primitives
func (f *ChannelConfigFromOrdererProviderFactory) CreateInfraProvider(config fab.EndpointConfig) (fab.InfraProvider, error) {

	fabProvider := fabpvdr.New(config)

	cfp := CustomInfraProvider{
		InfraProvider: fabProvider,
		orderer:       f.orderer,
	}
	return &cfp, nil
}

func setupOrderer(t *testing.T, endPointConfig fab.EndpointConfig, address string) fab.Orderer {

	//Get orderer config by orderer address
	oCfg, ok, _ := endPointConfig.OrdererConfig(resolveOrdererAddress(address))
	require.True(t, ok)

	o, err := orderer.New(endPointConfig, orderer.FromOrdererConfig(oCfg))
	require.Nil(t, err)

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
