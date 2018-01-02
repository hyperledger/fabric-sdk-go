// +build !pkcs11

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
)

func TestCryptoSuiteByConfigPKCS11Unsupported(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("was supposed to panic")
		}
	}()

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	//Prepare Config
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")

	//Get cryptosuite using config
	GetSuiteByConfig(mockConfig)
	t.Fatalf("Getting cryptosuite with unsupported pkcs11 security provider supposed to panic")
}
