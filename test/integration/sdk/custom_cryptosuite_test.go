/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"testing"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspSw "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

func customCryptoSuiteInit(t *testing.T) (*integration.BaseSetupImpl, string) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup

	//testSetup := integration.BaseSetupImpl{
	//	ConfigFile:    "../" + integration.ConfigTestFile,
	//	ChannelID:     "mychannel",
	//	OrgID:         org1Name,
	//	ChannelConfig: path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
	//}

	// Create SDK setup for the integration tests
	//sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	//if err != nil {
	//	t.Fatalf("Failed to create new SDK: %s", err)
	//}
	//defer sdk.Close()

	//if err := testSetup.Initialize(sdk); err != nil {
	//	t.Fatal(err)
	//}

	chaincodeID := integration.GenerateRandomID()
	resp, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID)
	require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	require.NotEmpty(t, resp, "instantiate response should be populated")

	return testSetup, chaincodeID
}

func TestEndToEndForCustomCryptoSuite(t *testing.T) {

	testSetup, chainCodeID := customCryptoSuiteInit(t)

	configBackend, err := integration.ConfigBackend()

	cryptoConfig := cryptosuite.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Failed to get config backend: %s", err)
	}

	//Get Test BCCSP,
	// TODO Need to use external BCCSP here
	customBccspProvider := getTestBCCSP(cryptoConfig)

	// Create SDK setup with custom cryptosuite provider factory
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithCorePkg(&CustomCryptoSuiteProviderFactory{bccspProvider: customBccspProvider}))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	response, err := chClient.Query(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	t.Logf("*** QueryValue before invoke %s", response.Payload)

	// Check the Query value
	if string(response.Payload) != "200" {
		t.Fatal("channel client query operation failed, upexpected query value")
	}

}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
	bccspProvider bccsp.BCCSP
}

// CreateCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	c := wrapper.NewCryptoSuite(f.bccspProvider)
	return c, nil
}

func getTestBCCSP(config core.CryptoSuiteConfig) bccsp.BCCSP {
	opts := getOptsByConfig(config)
	s, err := getBCCSPFromOpts(opts)
	if err != nil {
		panic(fmt.Sprintf("Failed getting software-based BCCSP [%s]", err))
	}

	return s
}

func getBCCSPFromOpts(config *bccspSw.SwOpts) (bccsp.BCCSP, error) {
	f := &bccspSw.SWFactory{}

	return f.Get(config)
}

func getOptsByConfig(c core.CryptoSuiteConfig) *bccspSw.SwOpts {
	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
	}

	return opts
}

/* TODO
func TestCustomCryptoSuite(t *testing.T) {
	testSetup := integration.BaseSetupImpl{
		ConfigFile: "../" + integration.ConfigTestFile,
	}

	defaultConfig, err := testSetup.InitConfig()()

	if err != nil {
		panic(fmt.Sprintf("Failed to get default config [%s]", err))
	}

	//Get Test BCCSP,
	customBccspProvider := getTestBCCSP(defaultConfig)
	//Get BCCSP custom wrapper for Test BCCSP
	customBccspWrapper := getBCCSPWrapper(customBccspProvider)

	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile),
		fabsdk.WithCorePkg(&CustomCryptoSuiteProviderFactory{bccspProvider: customBccspWrapper}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	key, err := sdk.CryptoSuiteProvider().KeyGen(nil)
	if err != nil {
		t.Fatalf("Failed to get key from  sdk.CryptoSuiteProvider().KeyGen(): %s", err)
	}

	bytes, err := key.Bytes()
	if err != nil {
		t.Fatalf("Failed to get key bytes from  sdk.CryptoSuiteProvider().KeyGen(): %s", err)
	}

	if string(bytes) != samplekey {
		t.Fatalf("Unexpected sdk.CryptoSuiteProvider(), expected to find BCCSPWrapper features : %s", err)
	}
}
*/
