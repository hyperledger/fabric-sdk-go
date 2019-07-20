/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/stretchr/testify/assert"
)

const (
	sampleMatchersOverrideAll    = "matchers_sample1.yaml"
	sampleMatchersRegexReplace   = "matchers_sample3.yaml"
	sampleMatchersIgnoreEndpoint = "matchers_sample6.yaml"
	sampleMatchersDir            = "matcher-samples"

	actualCAURL    = "https://ca.org1.example.com:7054"
	overridedCAURL = "https://ca.org1.example.com:8888"

	actualTargetServerName    = "ca.org1.example.com"
	overridedTargetServerName = "ca.override.example.com"
)

//TestCAURLOverride
//Scenario: Using entity mather to override CA URL
func TestCAURLOverride(t *testing.T) {

	//Test basic entity matcher
	matcherPath := filepath.Join(getConfigPath(), sampleMatchersDir, sampleMatchersOverrideAll)
	testCAEntityMatcher(t, matcherPath)

	//Test entity matcher with regex replace feature '$'
	matcherPath = filepath.Join(getConfigPath(), sampleMatchersDir, sampleMatchersRegexReplace)
	testCAEntityMatcher(t, matcherPath)
}

func testCAEntityMatcher(t *testing.T, configPath string) {
	//Without entity matcher
	configTestPath := filepath.Join(getConfigPath(), configTestFile)
	backends, err := getBackendsFromFiles(configTestPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 1, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok := config.CAConfig("ca.org1.example.com")
	assert.True(t, ok, "supposed to find caconfig")
	assert.NotNil(t, caConfig)
	assert.Equal(t, actualCAURL, caConfig.URL)
	assert.Equal(t, actualTargetServerName, caConfig.GRPCOptions["ssl-target-name-override"])

	//Using entity matcher to override CA URL
	backends, err = getBackendsFromFiles(configPath, configTestPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err = ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok = config.CAConfig("ca.org1.example.com")
	assert.True(t, ok, "supposed to find caconfig")
	assert.NotNil(t, caConfig)
	assert.Equal(t, overridedCAURL, caConfig.URL)
	assert.Equal(t, overridedTargetServerName, caConfig.GRPCOptions["ssl-target-name-override"])
}

//TestCAEntityMatcherIgnoreEndpoint tests CA entity matcher 'IgnoreEndpoint' option
// If marked 'IgnoreEndpoint: true' then corresponding CA will be ignored
func TestCAEntityMatcherIgnoreEndpoint(t *testing.T) {
	//Without entity matcher
	configTestPath := filepath.Join(getConfigPath(), configTestFile)
	matcherPath := filepath.Join(getConfigPath(), sampleMatchersDir, sampleMatchersIgnoreEndpoint)
	backends, err := getBackendsFromFiles(matcherPath, configTestPath)
	assert.Nil(t, err, "not supposed to get error")
	assert.Equal(t, 2, len(backends))

	config, err := ConfigFromBackend(backends...)
	assert.Nil(t, err, "not supposed to get error")
	assert.NotNil(t, config)

	caConfig, ok := config.CAConfig("ca.org1.example.com")
	assert.True(t, ok)
	assert.NotNil(t, caConfig)
	caConfig, ok = config.CAConfig("ca.org2.example.com")
	assert.False(t, ok)
	assert.Nil(t, caConfig)

	configImpl := config.(*IdentityConfig)
	assert.Equal(t, 1, len(configImpl.caConfigs))
	_, ok = configImpl.caConfigs["ca.org1.example.com"]
	assert.True(t, ok)
	_, ok = configImpl.caConfigs["ca.org2.example.com"]
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
