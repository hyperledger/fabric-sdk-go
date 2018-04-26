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
	configPath        = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_test.yaml"
	entityMangerLocal = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test//fixtures/config/overrides/local_entity_matchers.yaml"
	//LocalOrdererPeersCAsConfig config file to override on local test having only peers, orderers and CA entity entries
	LocalOrdererPeersCAsConfig = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/overrides/local_orderers_peers_ca.yaml"
	//LocalOrdererPeersConfig config file to override on local test having only peers and orderers entity entries
	LocalOrdererPeersConfig = "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/overrides/local_orderers_peers.yaml"
)

// ConfigBackend contains config backend for integration tests
var ConfigBackend = fetchConfigBackend(configPath, LocalOrdererPeersCAsConfig)

func fetchConfigBackend(configPath string, localOverride string) core.ConfigProvider {
	configProvider := config.FromFile(pathvar.Subst(configPath))

	args := os.Args[1:]
	for _, arg := range args {
		//If testlocal is enabled, then update config backend to run 'local' test
		if arg == "testLocal=true" {
			return func() ([]core.ConfigBackend, error) {
				return appendLocalEntityMappingBackend(configProvider, localOverride)
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
func AddLocalEntityMapping(configProvider core.ConfigProvider, configOverridePath string) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		return appendLocalEntityMappingBackend(configProvider, configOverridePath)
	}
}

//appendLocalEntityMappingBackend appends entity matcher backend to given config provider
func appendLocalEntityMappingBackend(configProvider core.ConfigProvider, configOverridePath string) ([]core.ConfigBackend, error) {
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

	//Local orderer/peer/CA config overrides
	configProvider = config.FromFile(pathvar.Subst(configOverridePath))
	localBackends, err := configProvider()
	if err != nil {
		return nil, err
	}

	//backends should fal back in this order - matcherBackends, localBackends, currentBackends
	localBackends = append(localBackends, matcherBackends...)
	localBackends = append(localBackends, currentBackends...)

	return localBackends, nil
}
