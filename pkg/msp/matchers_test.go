/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/stretchr/testify/assert"
)

const (
	sampleMatchersOverrideAll    = "../core/config/testdata/matcher-samples/matchers_sample1.yaml"
	sampleMatchersRegexReplace   = "../core/config/testdata/matcher-samples/matchers_sample3.yaml"
	sampleMatchersIgnoreEndpoint = "../core/config/testdata/matcher-samples/matchers_sample6.yaml"

	actualCAURL    = "https://ca.org1.example.com:7054"
	overridedCAURL = "https://ca.org1.example.com:8888"

	actualTargetServerName    = "ca.org1.example.com"
	overridedTargetServerName = "ca.override.example.com"
)

//TestCAURLOverride
//Scenario: Using entity mather to override CA URL
func TestCAURLOverride(t *testing.T) {

	//Test basic entity matcher
	testCAEntityMatcher(t, sampleMatchersOverrideAll)

	//Test entity matcher with regex replace feature '$'
	testCAEntityMatcher(t, sampleMatchersRegexReplace)
}

func testCAEntityMatcher(t *testing.T, configPath string) {
	//Without entity matcher
	backends, err := getBackendsFromFiles(configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 1, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok := config.CAConfig("org1")
	assert.True(t, ok, "supposed to find caconfig")
	assert.NotNil(t, caConfig)
	assert.Equal(t, actualCAURL, caConfig.URL)
	assert.Equal(t, actualTargetServerName, caConfig.GRPCOptions["ssl-target-name-override"])

	//Using entity matcher to override CA URL
	backends, err = getBackendsFromFiles(configPath, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err = ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok = config.CAConfig("org1")
	assert.True(t, ok, "supposed to find caconfig")
	assert.NotNil(t, caConfig)
	assert.Equal(t, overridedCAURL, caConfig.URL)
	assert.Equal(t, overridedTargetServerName, caConfig.GRPCOptions["ssl-target-name-override"])
}

//TestCAEntityMatcherIgnoreEndpoint tests CA entity matcher 'IgnoreEndpoint' option
// If marked 'IgnoreEndpoint: true' then corresponding CA will be ignored
func TestCAEntityMatcherIgnoreEndpoint(t *testing.T) {
	//Without entity matcher
	backends, err := getBackendsFromFiles(sampleMatchersIgnoreEndpoint, configTestFilePath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok := config.CAConfig("org1")
	assert.True(t, ok)
	assert.NotNil(t, caConfig)
	caConfig, ok = config.CAConfig("org2")
	assert.False(t, ok)
	assert.Nil(t, caConfig)

	configImpl := config.(*IdentityConfig)
	assert.Equal(t, 1, len(configImpl.caConfigsByOrg))
	_, ok = configImpl.caConfigsByOrg["org1"]
	assert.True(t, ok)
	_, ok = configImpl.caConfigsByOrg["org2"]
	assert.False(t, ok)
}

func getBackendsFromFiles(files ...string) ([]core.ConfigBackend, error) {

	backends := make([]core.ConfigBackend, len(files))
	for i, file := range files {
		backend, err := config.FromFile(file)()
		if err != nil {
			return nil, err
		}
		backends[i] = backend[0]
	}
	return backends, nil
}
