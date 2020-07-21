/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoConfigPrivKeyPathV1(t *testing.T) {
	const (
		cryptoConfigPath = "testdata/cryptoconfig/v1/{username}"
		username = "user"
	)
	ski := []byte{0,1}


	p := cryptoConfigPrivateKeyPath(cryptoConfigPath, username, ski)
	assert.Contains(t, p, "0001_sk")
}

func TestCryptoConfigPrivKeyPathV2(t *testing.T) {
	const (
		cryptoConfigRelPath = "testdata/cryptoconfig/v2/{username}"
		username = "user"
	)
	ski := []byte{0,1}

	cryptoConfigPath := filepath.Join(testDir(), cryptoConfigRelPath)
	t.Log(cryptoConfigPath)

	p := cryptoConfigPrivateKeyPath(cryptoConfigPath, username, ski)
	assert.Contains(t, p, "priv_sk")
}

func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}
