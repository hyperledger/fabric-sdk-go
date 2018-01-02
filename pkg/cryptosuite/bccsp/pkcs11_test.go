// +build pkcs11

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

func TestCryptoSuiteByConfigPKCS11Failure(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	//Prepare Config
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)
	mockConfig.EXPECT().SecurityProviderLibPath().Return("")
	mockConfig.EXPECT().SecurityProviderLabel().Return("")
	mockConfig.EXPECT().SecurityProviderPin().Return("")
	mockConfig.EXPECT().SoftVerify().Return(true)

	//Get cryptosuite using config
	samplecryptoSuite, err := GetSuiteByConfig(mockConfig)
	utils.VerifyNotEmpty(t, err, "Supposed to get error on GetSuiteByConfig call : %s", err)
	utils.VerifyEmpty(t, samplecryptoSuite, "Not supposed to get valid cryptosuite")
}
