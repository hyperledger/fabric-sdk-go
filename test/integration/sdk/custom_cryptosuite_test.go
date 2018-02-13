/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspSw "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const samplekey = "sample-key"

func TestEndToEndForCustomCryptoSuite(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	chainCodeID := integration.GenerateRandomID()
	if err := integration.InstallAndInstantiateExampleCC(testSetup.SDK, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	defaultConfig, err := testSetup.InitConfig()()

	if err != nil {
		panic(fmt.Sprintf("Failed to get default config [%s]", err))
	}

	//Get Test BCCSP,
	// TODO Need to use external BCCSP here
	customBccspProvider := getTestBCCSP(defaultConfig)

	// Create SDK setup with custom cryptosuite provider factory
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile),
		fabsdk.WithCorePkg(&CustomCryptoSuiteProviderFactory{bccspProvider: customBccspProvider}))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewClient(fabsdk.WithUser("User1")).Channel(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	response, err := chClient.Query(chclient.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	t.Logf("*** QueryValue before invoke %s", response.Payload)

	// Check the Query value
	if string(response.Payload) != "200" {
		t.Fatalf("channel client query operation failed, upexpected query value")
	}

	// Release all channel client resources
	chClient.Close()

}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
	bccspProvider bccsp.BCCSP
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	c := wrapper.NewCryptoSuite(f.bccspProvider)
	return c, nil
}

func getTestBCCSP(config apiconfig.Config) bccsp.BCCSP {
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

func getOptsByConfig(c apiconfig.Config) *bccspSw.SwOpts {
	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
		Ephemeral: c.Ephemeral(),
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

/*
	BCCSP Wrapper for test
*/

func getBCCSPWrapper(bccsp bccsp.BCCSP) bccsp.BCCSP {
	return &bccspWrapper{bccsp}
}

func getBCCSPKeyWrapper(key bccsp.Key) bccsp.Key {
	return &bccspKeyWrapper{key}
}

type bccspWrapper struct {
	bccsp.BCCSP
}

func (mock *bccspWrapper) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	return getBCCSPKeyWrapper(nil), nil
}

type bccspKeyWrapper struct {
	bccsp.Key
}

func (mock *bccspKeyWrapper) Bytes() ([]byte, error) {
	return []byte("sample-key"), nil
}
