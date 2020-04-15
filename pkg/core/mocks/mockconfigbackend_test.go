/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func getConfigPath() string {
	return filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata")
}

func TestMockConfigBackend(t *testing.T) {

	configPath := filepath.Join(getConfigPath(), "config_test.yaml")
	mockBackend, err := BackendFromFile(configPath)
	if err != nil {
		t.Fatalf("Unexpected error reading config: %s", err)
	}

	v, ok := mockBackend.Get("client")
	assert.True(t, ok, "!ok")
	assert.NotNil(t, v, "client not found")
	assert.True(t, reflect.TypeOf(v) == reflect.TypeOf(map[string]interface{}{}), "wrong type")

	v, ok = mockBackend.Get("client.tlscerts.systemcertpool")
	assert.True(t, ok, "!ok")
	assert.True(t, reflect.TypeOf(v) == reflect.TypeOf(true), "wrong type")
	assert.False(t, v.(bool), "wrong value")

	mockBackend.Set("client.tlscerts.systemcertpool", true)
	v, ok = mockBackend.Get("client.tlscerts.systemcertpool")
	assert.True(t, ok, "!ok")
	assert.True(t, reflect.TypeOf(v) == reflect.TypeOf(true), "wrong type")
	assert.True(t, v.(bool), "wrong value")

	v, ok = mockBackend.Get("some.new.value")
	assert.False(t, ok, "should not be ok")
	mockBackend.Set("some.new.value", "Hello World")
	v, ok = mockBackend.Get("some.new.value")
	assert.True(t, ok, "!ok")
	assert.True(t, reflect.TypeOf(v) == reflect.TypeOf(""), "wrong type")
	assert.Equal(t, "Hello World", v)

}
