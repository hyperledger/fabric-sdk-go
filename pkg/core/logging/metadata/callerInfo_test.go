/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
)

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := CallerInfo{}
	samppleModuleName := "sample-module-name"

	//By default caller info should be enabled if not set
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.DEBUG), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.INFO), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.WARNING), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.ERROR), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.CRITICAL), "Callerinfo supposed to be enabled for this level")

	sampleCallerInfoSetting.ShowCallerInfo(samppleModuleName, api.DEBUG)
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.DEBUG), "Callerinfo supposed to be enabled for this level")

	sampleCallerInfoSetting.HideCallerInfo(samppleModuleName, api.DEBUG)
	assert.False(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleModuleName, api.DEBUG), "Callerinfo supposed to be disabled for this level")

	//Reset existing caller info setting
	sampleCallerInfoSetting.showcaller = nil

	//By default caller info should be enabled for any module name
	samppleInvalidModuleName := "sample-module-name-doesnt-exists"
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, api.INFO), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, api.WARNING), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, api.ERROR), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, api.CRITICAL), "Callerinfo supposed to be enabled for this level")
	assert.True(t, sampleCallerInfoSetting.IsCallerInfoEnabled(samppleInvalidModuleName, api.DEBUG), "Callerinfo supposed to be enabled for this level")
}
