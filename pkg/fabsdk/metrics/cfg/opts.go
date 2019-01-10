/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cfg

import "github.com/pkg/errors"

// OperationsConfigOptions represents MetricsConfig interface with overridable interface functions
// if a function is not overridden, the default MetricsConfig implementation will be used.
type OperationsConfigOptions struct {
	operation
	metricCfg
}

type applier func()
type predicate func() bool
type setter struct{ isSet bool }

// operation interface allows to uniquely override MetricsConfig interface's OperationConfig() function
type operation interface {
	OperationCfg() OperationConfig
}

// caConfig interface allows to uniquely override MetricsConfig interface's CAConfig() function
type metricCfg interface {
	MetricCfg() MetricConfig
}

// BuildConfigMetricsFromOptions will return a MetricsConfig instance pre-built with Optional interfaces
// provided in fabsdk's WithMetricsConfig(opts...) call
func BuildConfigMetricsFromOptions(opts ...interface{}) (MetricsConfig, error) {
	// build a new MetricsConfig with overridden function implementations
	c := &OperationsConfigOptions{}
	for _, option := range opts {
		err := setMetricsConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the MetricsConfig were not updated with fabsdk's
// WithConfigMetrics(opts...) call, then use default MetricsConfig interface for these functions instead
func UpdateMissingOptsWithDefaultConfig(c *OperationsConfigOptions, d MetricsConfig) MetricsConfig {
	s := &setter{}

	s.set(c.operation, nil, func() { c.operation = d })
	s.set(c.metricCfg, nil, func() { c.metricCfg = d })

	return c
}

// IsMetricsConfigFullyOverridden will return true if all of the argument's sub interfaces is not nil
// (ie MetricsConfig interface not fully overridden)
func IsMetricsConfigFullyOverridden(c *OperationsConfigOptions) bool {
	return !anyNil(c.operation, c.metricCfg)
}

// will override MetricsConfig interface with functions provided by o (option)
func setMetricsConfigWithOptionInterface(c *OperationsConfigOptions, o interface{}) error {
	s := &setter{}

	s.set(c.operation, func() bool { _, ok := o.(operation); return ok }, func() { c.operation = o.(operation) })
	s.set(c.metricCfg, func() bool { _, ok := o.(metricCfg); return ok }, func() { c.metricCfg = o.(metricCfg) })

	if !s.isSet {
		return errors.Errorf("option %#v is not a sub interface of MetricsConfig, at least one of its functions must be implemented.", o)
	}
	return nil
}

// needed to avoid meta-linter errors (too many if conditions)
func (o *setter) set(current interface{}, check predicate, apply applier) {
	if current == nil && (check == nil || check()) {
		apply()
		o.isSet = true
	}
}

// will verify if any of objs element is nil, also needed to avoid meta-linter errors
func anyNil(objs ...interface{}) bool {
	for _, p := range objs {
		if p == nil {
			return true
		}
	}
	return false
}
