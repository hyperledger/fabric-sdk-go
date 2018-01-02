/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

func TestBadConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatalf("Unknown security provider should return error")
	}
}

func TestCryptoSuiteByConfigSW(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("")
	mockConfig.EXPECT().Ephemeral().Return(true)

	//Get cryptosuite using config
	c, err := GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}

	verifyHashFn(t, c)
}

func TestCryptoSuiteByBadConfigSW(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA0")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("")
	mockConfig.EXPECT().Ephemeral().Return(true)

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatalf("Bad configuration should return error")
	}
}

func TestCryptoSuiteDefaultEphemeral(t *testing.T) {
	c, err := GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}
	verifyHashFn(t, c)
}

func verifyHashFn(t *testing.T, c apicryptosuite.CryptoSuite) {
	msg := []byte("Hello")
	e := sha256.Sum256(msg)
	a, err := c.Hash(msg, &bccsp.SHA256Opts{})
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}

	if bytes.Compare(a, e[:]) != 0 {
		t.Fatalf("Expected SHA 256 hash function")
	}
}
