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

func TestLogLevels(t *testing.T) {

	mlevel := moduleLeveled{}

	mlevel.SetLevel("module-xyz-info", apilogging.INFO)
	mlevel.SetLevel("module-xyz-debug", apilogging.DEBUG)
	mlevel.SetLevel("module-xyz-error", apilogging.ERROR)
	mlevel.SetLevel("module-xyz-warning", apilogging.WARNING)

	//Run info level checks
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", apilogging.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-info", apilogging.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", apilogging.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", apilogging.WARNING))

	//Run debug level checks
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", apilogging.INFO))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", apilogging.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", apilogging.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", apilogging.WARNING))

	//Run info level checks
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", apilogging.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", apilogging.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-error", apilogging.ERROR))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", apilogging.WARNING))

	//Run info level checks
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", apilogging.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", apilogging.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", apilogging.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", apilogging.WARNING))

	//Run default log level check --> which is info currently
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", apilogging.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-random-module", apilogging.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", apilogging.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", apilogging.WARNING))

}
