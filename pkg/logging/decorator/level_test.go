/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package decorator

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

func TestLogLevels(t *testing.T) {

	mlevel := ModuleLevels{}

	mlevel.SetLevel("module-xyz-info", api.INFO)
	mlevel.SetLevel("module-xyz-debug", api.DEBUG)
	mlevel.SetLevel("module-xyz-error", api.ERROR)
	mlevel.SetLevel("module-xyz-warning", api.WARNING)

	//Run info level checks
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", api.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-info", api.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", api.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-info", api.WARNING))

	//Run debug level checks
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", api.INFO))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", api.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", api.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-debug", api.WARNING))

	//Run info level checks
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", api.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", api.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-error", api.ERROR))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-error", api.WARNING))

	//Run info level checks
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", api.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-warning", api.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", api.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-warning", api.WARNING))

	//Run default log level check --> which is info currently
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", api.INFO))
	utils.VerifyFalse(t, mlevel.IsEnabledFor("module-xyz-random-module", api.DEBUG))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", api.ERROR))
	utils.VerifyTrue(t, mlevel.IsEnabledFor("module-xyz-random-module", api.WARNING))

}
