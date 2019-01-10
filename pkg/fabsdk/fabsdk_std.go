// +build !pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
)

func (sdk *FabricSDK) initMetrics(config *configs) {
	//disabled metrics for standard build
	sdk.system = operations.NewSystem(operations.Options{
		Metrics: operations.MetricsOptions{
			Provider: "disabled",
		},
		Version: "latest",
	},
	)

	sdk.clientMetrics = &metrics.ClientMetrics{} // empty channel ClientMetrics for standard build.
}
