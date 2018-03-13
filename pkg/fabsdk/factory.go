/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"
)

// PkgSuite provides the package factories that create clients and providers
type PkgSuite interface {
	Core() (sdkApi.CoreProviderFactory, error)
	MSP() (sdkApi.MSPProviderFactory, error)
	Service() (sdkApi.ServiceProviderFactory, error)
	Logger() (api.LoggerProvider, error)
}
