/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

const (
	configPath           = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_test.yaml"
	configPathNoOrderer  = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_test_no_orderer.yaml"
	entityMangerLocal    = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test//fixtures/config/local_entity_matchers.yaml"
	localOrdererPeersCAs = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/local_orderers_peers_ca.yaml"
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
			return func() ([]core.ConfigBackend, error) {
				return appendLocalEntityMappingBackend(configProvider)
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
	return func() ([]core.ConfigBackend, error) {
		return appendLocalEntityMappingBackend(configProvider)
	}
}

//appendLocalEntityMappingBackend appends entity matcher backend to given config provider
func appendLocalEntityMappingBackend(configProvider core.ConfigProvider) ([]core.ConfigBackend, error) {
	//Current backend
	currentBackends, err := configProvider()
	if err != nil {
		return nil, err
	}

	//Entity matcher config backend
	configProvider = config.FromFile(pathvar.Subst(entityMangerLocal))
	matcherBackends, err := configProvider()
	if err != nil {
		return nil, err
	}

	//Local orderer, peer, CA config backend
	configProvider = config.FromFile(pathvar.Subst(localOrdererPeersCAs))
	localBackends, err := configProvider()
	if err != nil {
		return nil, err
	}

	//backends should fal back in this order - matcherBackends, localBackends, currentBackends
	localBackends = append(localBackends, matcherBackends...)
	localBackends = append(localBackends, currentBackends...)

	return localBackends, nil
}
