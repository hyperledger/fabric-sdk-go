/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cfg

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/pkg/errors"
)

// MetricsConfig contains operations system and metrics configuration
type MetricsConfig interface {
	OperationCfg() OperationConfig
	MetricCfg() MetricConfig
}

// OperationConfig defines an operations system configuration
type OperationConfig struct {
	TLSEnabled         bool
	ClientAuthRequired bool
	ListenAddress      string

	// TODO replace TLSCertFile, TLCKeyFile and ClientRootCAs to TLSCACerts (here and in the configs as well)
	TLSCertFile   string
	TLSKeyFile    string
	ClientRootCAs []string
	// END TODO
}

// MetricConfig defines a metric configuration used along the operation system config
type MetricConfig struct {
	// Provider : statsd, prometheus, or disabled
	Provider string
	// Statsd represents metrics config for Statsd provider
	Statsd Statsd
}

// Statsd config useful for Statsd metrics provider
type Statsd struct {
	// Network is statsd network type: tcp or udp
	Network string

	// Address is statsd server address: 127.0.0.1:8125
	Address string

	// WriteInterval is the interval at which locally cached counters and gauges are pushed
	// to statsd; timings are pushed immediately
	WriteInterval time.Duration

	// prefix is prepended to all emitted statsd metrics
	Prefix string
}

// MetricsConfigImpl is the default implementation of MetricsConfig holding config data loaded from backend
type MetricsConfigImpl struct {
	backend *lookup.ConfigLookup
	OperationConfig
	MetricConfig
}

// OperationCfg returns the oeprations Config
func (m *MetricsConfigImpl) OperationCfg() OperationConfig {
	return m.OperationConfig
}

// MetricCfg returns the metrics Config
func (m *MetricsConfigImpl) MetricCfg() MetricConfig {
	return m.MetricConfig
}

func (m *MetricsConfigImpl) loadMetricsConfiguration() error {
	m.createOperationCfg()
	err := m.createMetricCfg()
	if err != nil {
		return errors.WithMessage(err, "metric configuration load failed")
	}
	return nil
}

func (m *MetricsConfigImpl) createOperationCfg() {
	m.OperationConfig = OperationConfig{
		ListenAddress:      m.backend.GetString("operations.listenAddress"),
		ClientAuthRequired: m.backend.GetBool("operations.tls.clientAuthRequired"),
		TLSCertFile:        m.backend.GetString("operations.tls.cert.file"),
		TLSKeyFile:         m.backend.GetString("operations.tls.key.file"),
		TLSEnabled:         m.backend.GetBool("operations.tls.enabled"),
	}
	if rootCAs, ok := m.backend.Lookup("operations.tls.clientRootCAs.files"); ok {
		rootCAStrs := make([]string, len(rootCAs.([]interface{})))
		for i, r := range rootCAs.([]interface{}) {
			rootCAStrs[i] = r.(string)
		}
		m.OperationConfig.ClientRootCAs = rootCAStrs
	}

}

func (m *MetricsConfigImpl) createMetricCfg() error {
	m.MetricConfig = MetricConfig{
		Provider: m.backend.GetString("metrics.provider"),
	}

	err := m.backend.UnmarshalKey("metrics.statsd", &m.MetricConfig.Statsd)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'metric.statsd' config item to MetricConfig.Statsd type")
	}

	return nil
}

//ConfigFromBackend returns identity config implementation of given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) (MetricsConfig, error) {
	//create default metrics config
	config := &MetricsConfigImpl{backend: lookup.New(coreBackend...)}
	//operationsConfig
	if err := config.loadMetricsConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "metrics configuration load failed")
	}
	return config, nil
}
