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
	configPath = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/config/config_test.yaml"
	//entityMatcherLocal config file containing entity matchers for local test
	entityMatcherLocal = "${FABRIC_SDK_GO_PROJECT_PATH}/test//fixtures/config/overrides/local_entity_matchers.yaml"
	//ConfigPathSingleOrg single org version of 'configPath' for testing discovery
	ConfigPathSingleOrg = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/config/config_e2e_single_org.yaml"
)

// ConfigBackend contains config backend for integration tests
var ConfigBackend = fetchConfigBackend(configPath, entityMatcherLocal)

// fetchConfigBackend returns a ConfigProvider that retrieves config data from the given configPath,
// or from the given overrides for local testing
func fetchConfigBackend(configPath string, entityMatcherOverride string) core.ConfigProvider {
	configProvider := config.FromFile(pathvar.Subst(configPath))

	if IsLocal() {
		return func() ([]core.ConfigBackend, error) {
			return appendLocalEntityMappingBackend(configProvider, entityMatcherOverride)
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
		return appendLocalEntityMappingBackend(configProvider, entityMatcherLocal)
	}
}

func extractBackend(configProvider core.ConfigProvider) ([]core.ConfigBackend, error) {
	if configProvider == nil {
		return []core.ConfigBackend{}, nil
	}
	return configProvider()
}

//appendLocalEntityMappingBackend appends entity matcher backend to given config provider
func appendLocalEntityMappingBackend(configProvider core.ConfigProvider, entityMatcherOverridePath string) ([]core.ConfigBackend, error) {
	currentBackends, err := extractBackend(configProvider)
	if err != nil {
		return nil, err
	}

	//Entity matcher config backend
	configProvider = config.FromFile(pathvar.Subst(entityMatcherOverridePath))
	matcherBackends, err := configProvider()
	if err != nil {
		return nil, err
	}

	//backends should fal back in this order - matcherBackends, localBackends, currentBackends
	localBackends := append([]core.ConfigBackend{}, matcherBackends...)
	localBackends = append(localBackends, currentBackends...)

	return localBackends, nil
}

//IsDynamicDiscoverySupported returns if fabric version on which tests are running supports dynamic discovery
//any version greater than v1.1 supports dynamic discovery
func IsDynamicDiscoverySupported() bool {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "fabric-fixture=v1.1" {
			//not supported for fabric fixture v1.1
			return false
		}
	}
	return true
}
