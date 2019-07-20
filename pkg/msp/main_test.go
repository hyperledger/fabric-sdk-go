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
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	org1               = "Org1"
	org1CA             = "ca.org1.example.com"
	org2CA             = "ca.org2.example.com"
	caServerURLListen  = "http://127.0.0.1:0"
	dummyUserStorePath = "/tmp/userstore"
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

func (f *textFixture) setup(configBackend ...core.ConfigBackend) { //nolint

	if len(configBackend) == 0 {
		configPath := filepath.Join(getConfigPath(), configTestFile)
		backend, err := getCustomBackend(configPath)
		if err != nil {
			panic(err)
		}
		configBackend = backend
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

	configBackend = updateCAServerURL(caServerURL, configBackend)

	f.cryptSuiteConfig = cryptosuite.ConfigFromBackend(configBackend...)

	f.endpointConfig, err = fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config : %s", err))
	}

	f.identityConfig, err = ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config : %s", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	f.cryptoSuite, err = sw.GetSuiteByConfig(f.cryptSuiteConfig)
	if f.cryptoSuite == nil {
		panic(fmt.Sprintf("Failed initialize cryptoSuite: %s", err))
	}

	if f.identityConfig.CredentialStorePath() != "" {
		f.userStore, err = NewCertFileUserStore(f.identityConfig.CredentialStorePath())
		if err != nil {
			panic(fmt.Sprintf("creating a user store failed: %s", err))
		}
	}
	f.userStore = userStoreFromConfig(nil, f.identityConfig)

	identityManagers := make(map[string]msp.IdentityManager)
	netConfig := f.endpointConfig.NetworkConfig()
	if netConfig == nil {
		panic("failed to get network config")
	}
	for orgName := range netConfig.Organizations {
		mgr, err1 := NewIdentityManager(orgName, f.userStore, f.cryptoSuite, f.endpointConfig)
		if err1 != nil {
			panic(fmt.Sprintf("failed to initialize identity manager for organization: %s, cause :%s", orgName, err1))
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
		panic(fmt.Sprintf("failed to created context for test setup: %s", err))
	}

	f.caClient, err = NewCAClient(org1, ctx)
	if err != nil {
		panic(fmt.Sprintf("NewCAClient returned error: %s", err))
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
	cert, err := ioutil.ReadFile(filepath.Join("testdata", "root.pem"))
	if err != nil {
		t.Fatalf("Error reading cert: %s", err)
	}
	return cert
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %s\n", storePath, err))
	}
}

func cleanupTestPath(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %s", storePath, err)
	}
}

func mspIDByOrgName(t *testing.T, c fab.EndpointConfig, orgName string) string {
	netConfig := c.NetworkConfig()
	if netConfig == nil {
		t.Fatal("network config retrieval failed")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatal("org config retrieval failed ")
	}
	return orgConfig.MSPID
}

func userStoreFromConfig(t *testing.T, config msp.IdentityConfig) msp.UserStore {
	csp := config.CredentialStorePath()
	if csp == "" {
		return nil
	}
	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: csp})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %s", err)
	}
	userStore, err := NewCertFileUserStore1(stateStore)
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %s", err)
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

func updateCAServerURL(caServerURL string, existingBackends []core.ConfigBackend) []core.ConfigBackend {

	//get existing certificateAuthorities
	networkConfig := identityConfigEntity{}
	lookup.New(existingBackends...).UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)

	//update URLs
	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.URL = caServerURL

	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca2Config.URL = caServerURL

	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities[".ca.org2.example.com"] = ca2Config

	//update backend
	backendMap := make(map[string]interface{})
	//Override backend with updated certificate authorities config
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, existingBackends...)

	return backends
}
