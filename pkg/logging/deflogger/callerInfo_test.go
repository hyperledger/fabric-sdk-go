/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deflogger

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := callerInfo{}
	samppleModuleName := "sample-module-name"

	sampleCallerInfoSetting.ShowCallerInfo(samppleModuleName, apilogging.DEBUG)
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.DEBUG), "Callerinfo supposed to be enabled for this level")

	sampleCallerInfoSetting.HideCallerInfo(samppleModuleName, apilogging.DEBUG)
	utils.VerifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.DEBUG), "Callerinfo supposed to be disabled for this level")

	//Reset existing caller info setting
	sampleCallerInfoSetting.showcaller = nil

	//By default caller info should be disabled if not set
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.DEBUG), "Callerinfo supposed to be enabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.INFO), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.WARNING), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.ERROR), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, apilogging.CRITICAL), "Callerinfo supposed to be disabled for this level")

	//By default caller info should be disabled if module name not found
	samppleInvalidModuleName := "sample-module-name-doesnt-exists"
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, apilogging.INFO), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, apilogging.WARNING), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, apilogging.ERROR), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, apilogging.CRITICAL), "Callerinfo supposed to be disabled for this level")
	utils.VerifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, apilogging.DEBUG), "Callerinfo supposed to be disabled for this level")
}
