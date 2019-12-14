/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package operations

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/metrics"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/metrics/prometheus"
)

var (
	fabricVersion = metrics.GaugeOpts{
		Name:         "fabric_version",
		Help:         "The active version of Fabric.",
		LabelNames:   []string{"version"},
		StatsdFormat: "%{#fqname}.%{version}",
	}

	gaugeLock        sync.Mutex
	promVersionGauge metrics.Gauge
)

func versionGauge(provider metrics.Provider) metrics.Gauge {
	switch provider.(type) {
	case *prometheus.Provider:
		gaugeLock.Lock()
		defer gaugeLock.Unlock()
		if promVersionGauge == nil {
			promVersionGauge = provider.NewGauge(fabricVersion)
		}
		return promVersionGauge

	default:
		return provider.NewGauge(fabricVersion)
	}
}
