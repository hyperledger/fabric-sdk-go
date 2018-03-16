/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package metadata

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/stretchr/testify/assert"
)

func TestLogLevels(t *testing.T) {

	mlevel := ModuleLevels{}

	mlevel.SetLevel("module-xyz-info", api.INFO)
	mlevel.SetLevel("module-xyz-debug", api.DEBUG)
	mlevel.SetLevel("module-xyz-error", api.ERROR)
	mlevel.SetLevel("module-xyz-warning", api.WARNING)

	//Run info level checks
	assert.True(t, mlevel.IsEnabledFor("module-xyz-info", api.INFO))
	assert.False(t, mlevel.IsEnabledFor("module-xyz-info", api.DEBUG))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-info", api.ERROR))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-info", api.WARNING))

	//Run debug level checks
	assert.True(t, mlevel.IsEnabledFor("module-xyz-debug", api.INFO))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-debug", api.DEBUG))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-debug", api.ERROR))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-debug", api.WARNING))

	//Run info level checks
	assert.False(t, mlevel.IsEnabledFor("module-xyz-error", api.INFO))
	assert.False(t, mlevel.IsEnabledFor("module-xyz-error", api.DEBUG))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-error", api.ERROR))
	assert.False(t, mlevel.IsEnabledFor("module-xyz-error", api.WARNING))

	//Run info level checks
	assert.False(t, mlevel.IsEnabledFor("module-xyz-warning", api.INFO))
	assert.False(t, mlevel.IsEnabledFor("module-xyz-warning", api.DEBUG))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-warning", api.ERROR))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-warning", api.WARNING))

	//Run default log level check --> which is info currently
	assert.True(t, mlevel.IsEnabledFor("module-xyz-random-module", api.INFO))
	assert.False(t, mlevel.IsEnabledFor("module-xyz-random-module", api.DEBUG))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-random-module", api.ERROR))
	assert.True(t, mlevel.IsEnabledFor("module-xyz-random-module", api.WARNING))

}
