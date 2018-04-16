/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"strings"

	"regexp"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/spf13/viper"
)

const (
	configPath          = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_test.yaml"
	configPathNoOrderer = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_test_no_orderer.yaml"
	entityMangerLocal   = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test//fixtures/config/entity_matchers_local.yaml"
)

// ConfigBackend contains config backend for integration tests
var ConfigBackend = fetchConfigBackend(configPath)

// ConfigNoOrdererBackend contains config backend for no orderer integration tests
var ConfigNoOrdererBackend = fetchConfigBackend(configPathNoOrderer)

func fetchConfigBackend(configPath string) core.ConfigProvider {
	configProvider := config.FromFile(pathvar.Subst(configPath))

	args := os.Args[1:]
	for _, arg := range args {
		//If testlocal is enabled, then update config backend to run 'local' test
		if arg == "testLocal=true" {
			return func() (core.ConfigBackend, error) {
				configBackend, err := configProvider()
				if err != nil {
					return nil, err
				}
				return addLocalEntityMappingToBackend(configBackend)
			}
		}
	}

	return configProvider
}

//IsLocal checks os argument and returns true if 'testLocal=true' argument found
func IsLocal() bool {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "testLocal=true" {
			return true
		}
	}
	return false
}

//AddLocalEntityMapping adds local test entity mapping to config backend
// and returns updated config provider
func AddLocalEntityMapping(configProvider core.ConfigProvider) core.ConfigProvider {
	backend, err := configProvider()
	if err != nil {
		return func() (core.ConfigBackend, error) {
			return nil, err
		}
	}
	return func() (core.ConfigBackend, error) {
		return addLocalEntityMappingToBackend(backend)
	}
}

//addLocalEntityMappingToBackend adds local test entity mapping to config backend
func addLocalEntityMappingToBackend(backend core.ConfigBackend) (core.ConfigBackend, error) {

	//Read entity matchers through viper and add it to config backend
	myViper := viper.New()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	myViper.SetConfigFile(pathvar.Subst(entityMangerLocal))
	err := myViper.MergeInConfig()
	if err != nil {
		return nil, err
	}
	backendMap := make(map[string]interface{})
	backendMap["entityMatchers"] = myViper.Get("entityMatchers")

	//TODO delete below lines once entity matchers issue is solved
	// temporary fix starts here
	re := regexp.MustCompile(".*:")
	networkConfig := fab.NetworkConfig{}
	configLookup := lookup.New(backend)

	if err = configLookup.UnmarshalKey("orderers", &networkConfig.Orderers); err != nil {
		return nil, err
	}
	if err = configLookup.UnmarshalKey("peers", &networkConfig.Peers); err != nil {
		return nil, err
	}
	orderer, ok := networkConfig.Orderers["local.orderer.example.com"]
	if ok {
		orderer.URL = re.ReplaceAllString(orderer.URL, "localhost:")
		networkConfig.Orderers["local.orderer.example.com"] = orderer
	}

	peer1, ok := networkConfig.Peers["local.peer0.org1.example.com"]
	if ok {
		peer1.URL = re.ReplaceAllString(peer1.URL, "localhost:")
		peer1.EventURL = re.ReplaceAllString(peer1.EventURL, "localhost:")
		networkConfig.Peers["local.peer0.org1.example.com"] = peer1
	}

	peer2, ok := networkConfig.Peers["local.peer0.org2.example.com"]
	if ok {
		peer2.URL = re.ReplaceAllString(peer2.URL, "localhost:")
		peer2.EventURL = re.ReplaceAllString(peer2.EventURL, "localhost:")
		networkConfig.Peers["local.peer0.org2.example.com"] = peer2
	}

	backendMap["orderers"] = networkConfig.Orderers
	backendMap["peers"] = networkConfig.Peers
	// temporary fix ends here

	return &mocks.MockConfigBackend{KeyValueMap: backendMap, CustomBackend: backend}, nil

}
