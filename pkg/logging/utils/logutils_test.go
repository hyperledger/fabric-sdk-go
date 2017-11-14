/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

func TestGetLogLevels(t *testing.T) {

	level, err := LogLevel("info")
	verifyLogLevel(t, apilogging.INFO, level, err, true)

	level, err = LogLevel("iNfO")
	verifyLogLevel(t, apilogging.INFO, level, err, true)

	level, err = LogLevel("debug")
	verifyLogLevel(t, apilogging.DEBUG, level, err, true)

	level, err = LogLevel("DeBuG")
	verifyLogLevel(t, apilogging.DEBUG, level, err, true)

	level, err = LogLevel("warning")
	verifyLogLevel(t, apilogging.WARNING, level, err, true)

	level, err = LogLevel("WarNIng")
	verifyLogLevel(t, apilogging.WARNING, level, err, true)

	level, err = LogLevel("error")
	verifyLogLevel(t, apilogging.ERROR, level, err, true)

	level, err = LogLevel("eRRoR")
	verifyLogLevel(t, apilogging.ERROR, level, err, true)

	level, err = LogLevel("outofthebox")
	verifyLogLevel(t, -1, level, err, false)
	//
	//level, err = LogLevel("")
	//verifyLogLevel(t, -1, level, err, false)
}

func verifyLogLevel(t *testing.T, expectedLevel apilogging.Level, currentlevel apilogging.Level, err error, success bool) {
	if success {
		VerifyEmpty(t, err, "not supposed to get error for this scenario")
	} else {
		VerifyNotEmpty(t, err, "supposed to get error for this scenario, but got error : %v", err)
		return
	}

	VerifyTrue(t, currentlevel == expectedLevel, "unexpected log level : expected '%s', but got '%s'", expectedLevel, currentlevel)
}
