/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics/cfg"

// metricconfig_override_test.go is an example of programmatically configuring the sdk by injecting instances that implement Metricsonfig's functions (representing the sdk's configs)
// for the sake of overriding MetricsConfig integration tests, the structure variables below set the metrics to disabled as the standarad build does not use metrics.
// Using the pprof build tag, application developers can create sub interfaces of MetricsConfig with values similar to what is found in /test/fixtures/config/config_test.yaml
// the example implementation functions in this file can be overridden to load configs in any way that suits the client application needs

var (
	operationConfig = cfg.OperationConfig{
		ListenAddress:      "127.0.0.1:8080",
		TLSEnabled:         false,
		TLSCertFile:        "",
		TLSKeyFile:         "",
		ClientAuthRequired: false,
		ClientRootCAs:      []string{},
	}

	metricConfig = cfg.MetricConfig{
		Provider: "disabled",
		Statsd:   cfg.Statsd{},
	}

	opConfigImpl          = &exampleOperation{}
	metricCfgImpl         = &exampleMetric{}
	operationsConfigImpls = []interface{}{
		opConfigImpl,
		metricCfgImpl,
	}
)

type exampleOperation struct{}

//OperationCfg overrides MetricsConfig's OperationConfig function which returns the operations system config
func (m *exampleOperation) OperationCfg() cfg.OperationConfig {
	return operationConfig
}

type exampleMetric struct{}

//MetricCfg overrides MetricsConfig's MetricConfig function which returns the metrics specific config
func (m *exampleMetric) MetricCfg() cfg.MetricConfig {
	return metricConfig
}
