/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/mocks"
)

const (
	org1                        = "Org1"
	caServerURL                 = "http://localhost:8090"
	dummyUserStorePath          = "/tmp/userstore"
	fullConfigPath              = "testdata/config_test.yaml"
	wrongURLConfigPath          = "testdata/config_wrong_url.yaml"
	noCAConfigPath              = "testdata/config_no_ca.yaml"
	embeddedRegistrarConfigPath = "testdata/config_embedded_registrar.yaml"
	noRegistrarConfigPath       = "testdata/config_no_registrar.yaml"
)

type textFixture struct {
	config          core.Config
	cryptoSuite     core.CryptoSuite
	stateStore      core.KVStore
	userStore       msp.UserStore
	identityManager *IdentityManager
	caClient        mspapi.CAClient
}

var caServer = &mocks.MockFabricCAServer{}

func (f *textFixture) setup(configPath string) {

	if configPath == "" {
		configPath = fullConfigPath
	}

	var err error
	f.config, err = config.FromFile(configPath)()
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.config.KeyStorePath())
	cleanup(f.config.CredentialStorePath())

	f.cryptoSuite, err = sw.GetSuiteByConfig(f.config)
	if f.cryptoSuite == nil {
		panic(fmt.Sprintf("Failed initialize cryptoSuite: %v", err))
	}
	if f.config.CredentialStorePath() != "" {
		f.userStore, err = NewCertFileUserStore(f.config.CredentialStorePath())
		if err != nil {
			panic(fmt.Sprintf("creating a user store failed: %v", err))
		}
	}
	f.stateStore = stateStoreFromConfig(nil, f.config)

	f.identityManager, err = NewIdentityManager("org1", f.stateStore, f.cryptoSuite, f.config)
	if err != nil {
		panic(fmt.Sprintf("manager.NewManager returned error: %v", err))
	}

	f.caClient, err = NewCAClient(org1, f.identityManager, f.stateStore, f.cryptoSuite, f.config)
	if err != nil {
		panic(fmt.Sprintf("NewCAClient returned error: %v", err))
	}

	// Start Http Server if it's not running
	caServer.Start(strings.TrimPrefix(caServerURL, "http://"), f.cryptoSuite)
}

func (f *textFixture) close() {
	cleanup(f.config.CredentialStorePath())
	cleanup(f.config.KeyStorePath())
}

// readCert Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
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

func mspIDByOrgName(t *testing.T, c core.Config, orgName string) string {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("org config retrieval failed: %v", err)
	}
	return orgConfig.MspID
}

func stateStoreFromConfig(t *testing.T, config core.Config) core.KVStore {
	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: config.CredentialStorePath()})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %v", err)
	}
	return stateStore
}
