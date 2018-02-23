/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package loglevel

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/testutils"
)

func TestLogLevels(t *testing.T) {

	mlevel := ModuleLevels{}

	mlevel.SetLevel("module-xyz-info", INFO)
	mlevel.SetLevel("module-xyz-debug", DEBUG)
	mlevel.SetLevel("module-xyz-error", ERROR)
	mlevel.SetLevel("module-xyz-warning", WARNING)

	//Run info level checks
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", INFO))
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-info", DEBUG))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", ERROR))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", WARNING))

	//Run debug level checks
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", INFO))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", DEBUG))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", ERROR))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", WARNING))

	//Run info level checks
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", INFO))
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", DEBUG))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-error", ERROR))
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", WARNING))

	//Run info level checks
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", INFO))
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", DEBUG))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", ERROR))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", WARNING))

	//Run default log level check --> which is info currently
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", INFO))
	testutils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-random-module", DEBUG))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", ERROR))
	testutils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", WARNING))

}
