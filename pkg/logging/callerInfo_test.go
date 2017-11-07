/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import "testing"

func TestCallerInfoSetting(t *testing.T) {

	sampleCallerInfoSetting := callerInfo{}
	sampleModuleName := "module-xyz-info"

	sampleCallerInfoSetting.ShowCallerInfo(sampleModuleName)
	verifyTrue(t, sampleCallerInfoSetting.IsCallerInfoEnabled(sampleModuleName), "Callerinfo supposed to be enabled for this module")

	sampleCallerInfoSetting.HideCallerInfo(sampleModuleName)
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(sampleModuleName), "Callerinfo supposed to be disabled for this module")

	//By default caller info should be disabled
	verifyFalse(t, sampleCallerInfoSetting.IsCallerInfoEnabled(sampleModuleName+"DUMMY"), "Callerinfo supposed to be disabled for this module")

}
