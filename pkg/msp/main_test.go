/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	org1                        = "Org1"
	caServerURLListen           = "http://127.0.0.1:0"
	dummyUserStorePath          = "/tmp/userstore"
	fullConfigPath              = "testdata/config_test.yaml"
	wrongURLConfigPath          = "testdata/config_wrong_url.yaml"
	noCAConfigPath              = "testdata/config_no_ca.yaml"
	embeddedRegistrarConfigPath = "testdata/config_embedded_registrar.yaml"
	noRegistrarConfigPath       = "testdata/config_no_registrar.yaml"
)

var caServerURL string

type textFixture struct {
	endpointConfig          fab.EndpointConfig
	identityConfig          msp.IdentityConfig
	cryptSuiteConfig        core.CryptoSuiteConfig
	cryptoSuite             core.CryptoSuite
	userStore               msp.UserStore
	caClient                mspapi.CAClient
	identityManagerProvider msp.IdentityManagerProvider
}

var caServer = &mockmsp.MockFabricCAServer{}

func (f *textFixture) setup(configPath string) {

	if configPath == "" {
		configPath = fullConfigPath
	}

	var lis net.Listener
	var err error
	if !caServer.Running() {
		lis, err = net.Listen("tcp", strings.TrimPrefix(caServerURLListen, "http://"))
		if err != nil {
			panic(fmt.Sprintf("Error starting CA Server %s", err))
		}

		caServerURL = "http://" + lis.Addr().String()
	}

	cfgRaw := readConfigWithReplacement(configPath, "http://localhost:8050", caServerURL)
	configBackend, err := config.FromRaw(cfgRaw, "yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read config backend: %v", err))
	}

	f.cryptSuiteConfig, f.endpointConfig, f.identityConfig, err = config.FromBackend(configBackend)()
	if err != nil {
		panic(fmt.Sprintf("Failed to read config : %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	f.cryptoSuite, err = sw.GetSuiteByConfig(f.cryptSuiteConfig)
	if f.cryptoSuite == nil {
		panic(fmt.Sprintf("Failed initialize cryptoSuite: %v", err))
	}

	if f.identityConfig.CredentialStorePath() != "" {
		f.userStore, err = NewCertFileUserStore(f.identityConfig.CredentialStorePath())
		if err != nil {
			panic(fmt.Sprintf("creating a user store failed: %v", err))
		}
	}
	f.userStore = userStoreFromConfig(nil, f.identityConfig)

	identityManagers := make(map[string]msp.IdentityManager)
	netConfig, err := f.endpointConfig.NetworkConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to get network config: %v", err))
	}
	for orgName := range netConfig.Organizations {
		mgr, err := NewIdentityManager(orgName, f.userStore, f.cryptoSuite, f.endpointConfig)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize identity manager for organization: %s, cause :%v", orgName, err))
		}
		identityManagers[orgName] = mgr
	}

	f.identityManagerProvider = &identityManagerProvider{identityManager: identityManagers}

	ctxProvider := context.NewProvider(context.WithIdentityManagerProvider(f.identityManagerProvider),
		context.WithUserStore(f.userStore), context.WithCryptoSuite(f.cryptoSuite),
		context.WithCryptoSuiteConfig(f.cryptSuiteConfig), context.WithEndpointConfig(f.endpointConfig),
		context.WithIdentityConfig(f.identityConfig))

	ctx := &context.Client{Providers: ctxProvider}

	if err != nil {
		panic(fmt.Sprintf("failed to created context for test setup: %v", err))
	}

	f.caClient, err = NewCAClient(org1, ctx)
	if err != nil {
		panic(fmt.Sprintf("NewCAClient returned error: %v", err))
	}

	// Start Http Server if it's not running
	if !caServer.Running() {
		caServer.Start(lis, f.cryptoSuite)
	}
}

func (f *textFixture) close() {
	cleanup(f.identityConfig.CredentialStorePath())
	cleanup(f.cryptSuiteConfig.KeyStorePath())
}

// readCert Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}

func readConfigWithReplacement(path string, origURL, newURL string) []byte {
	cfgRaw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config [%s]", err))
	}

	updatedCfg := strings.Replace(string(cfgRaw), origURL, newURL, -1)
	return []byte(updatedCfg)
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}

func cleanupTestPath(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}

func mspIDByOrgName(t *testing.T, c fab.EndpointConfig, orgName string) string {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("org config retrieval failed: %v", err)
	}
	return orgConfig.MSPID
}

func userStoreFromConfig(t *testing.T, config msp.IdentityConfig) msp.UserStore {
	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: config.CredentialStorePath()})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %v", err)
	}
	userStore, err := NewCertFileUserStore1(stateStore)
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %v", err)
	}
	return userStore
}

type identityManagerProvider struct {
	identityManager map[string]msp.IdentityManager
}

// IdentityManager returns the organization's identity manager
func (p *identityManagerProvider) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	im, ok := p.identityManager[strings.ToLower(orgName)]
	if !ok {
		return nil, false
	}
	return im, true
}
