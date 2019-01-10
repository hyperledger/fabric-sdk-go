/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cfg

import (
	"testing"
	"time"
)

var (
	m0 = &MetricsConfigImpl{}
	m1 = &mockOperationConfig{}
	m2 = &mockMetricConfig{}
)

func TestCreateCustomFullMetricsConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface implementation)
	metricConfigOption, err := BuildConfigMetricsFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildConfigMetricsFromOptions returned unexpected error %s", err)
	}
	if metricConfigOption == nil {
		t.Fatal("BuildConfigMetricsFromOptions call returned nil")
	}
}

func TestCreateCustomMetricConfig(t *testing.T) {
	// try to build with separate interfaces
	metricsConfigOption, err := BuildConfigMetricsFromOptions(m1, m2)
	if err != nil {
		t.Fatalf("BuildConfigMetricsFromOptions returned unexpected error %s", err)
	}
	var oco *OperationsConfigOptions
	var ok bool
	if oco, ok = metricsConfigOption.(*OperationsConfigOptions); !ok {
		t.Fatalf("BuildConfigMetricsFromOptions did not return a Options instance %T", metricsConfigOption)
	}
	if oco == nil {
		t.Fatal("build OperationsConfigOptions returned is nil")
	}

	metCfg := oco.MetricCfg()
	if &metCfg == nil {
		t.Fatalf("MetricsConfig was supposed to have MetricCfg function overridden from Options but was not %+v. MetricCfg: %s", oco, metCfg)
	}

	opCfg := oco.OperationCfg()
	if &opCfg == nil {
		t.Fatalf("MetricsConfig was supposed to have OperationCfg function overridden from Options but was not %+v. OperationCfg: %s", oco, metCfg)
	}

}

func TestIsMetricsConfigFullyOverridden(t *testing.T) {
	// test with the some interfaces
	metricsConfigOption, err := BuildConfigMetricsFromOptions(m1)
	if err != nil {
		t.Fatalf("BuildConfigMetricsFromOptions returned unexpected error %s", err)
	}

	var oco *OperationsConfigOptions
	var ok bool
	if oco, ok = metricsConfigOption.(*OperationsConfigOptions); !ok {
		t.Fatalf("BuildConfigMetricsFromOptions did not return a Options instance %T", metricsConfigOption)
	}

	// test verify if some interfaces were not overridden according to BuildConfigEndpointFromOptions above,
	// only 1 interface was overridden, so expected value is false
	isFullyOverridden := IsMetricsConfigFullyOverridden(oco)
	if isFullyOverridden {
		t.Fatal("Expected not fully overridden MetricsConfig interface, but received fully overridden.")
	}

	// now try with no opts, expected value is also false
	metricsConfigOption, err = BuildConfigMetricsFromOptions()
	if err != nil {
		t.Fatalf("BuildConfigMetricsFromOptions returned unexpected error %s", err)
	}
	if oco, ok = metricsConfigOption.(*OperationsConfigOptions); !ok {
		t.Fatalf("BuildConfigMetricsFromOptions did not return a Options instance %T", metricsConfigOption)
	}

	isFullyOverridden = IsMetricsConfigFullyOverridden(oco)
	if isFullyOverridden {
		t.Fatal("Expected not fully overridden MetricsConfig interface with empty options, but received fully overridden.")
	}

	// now try with all opts, expected value is true this time
	metricsConfigOption, err = BuildConfigMetricsFromOptions(m1, m2)
	if err != nil {
		t.Fatalf("BuildConfigMetricsFromOptions returned unexpected error %s", err)
	}
	if oco, ok = metricsConfigOption.(*OperationsConfigOptions); !ok {
		t.Fatalf("BuildConfigMetricsFromOptions did not return a Options instance %T", metricsConfigOption)
	}

	isFullyOverridden = IsMetricsConfigFullyOverridden(oco)
	if !isFullyOverridden {
		t.Fatal("Expected fully overridden MetricsConfig interface, but received not fully overridden.")
	}
}

type mockOperationConfig struct{}

func (m *mockOperationConfig) OperationCfg() OperationConfig {
	return OperationConfig{}
}

type mockMetricConfig struct{}

func (m *mockMetricConfig) MetricCfg() MetricConfig {
	return MetricConfig{
		Provider: "disabled",
		Statsd: Statsd{
			Prefix:        "test",
			WriteInterval: 2 * time.Second,
			Address:       "127.0.0.1:8080",
			Network:       "udp",
		},
	}
}
