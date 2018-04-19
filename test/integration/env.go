/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
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

	return &mocks.MockConfigBackend{KeyValueMap: backendMap, CustomBackend: backend}, nil

}
