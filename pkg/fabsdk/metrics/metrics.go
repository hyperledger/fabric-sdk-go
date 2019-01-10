/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metrics

import "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/metrics"

var (
	// for now, only channel clients require metrics tracking. TODO: update to generalize metrics for other client types if needed.
	queriesReceived = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "queries_received",
		Help:         "The number of channel client queries received.",
		LabelNames:   []string{"chaincode", "Fcn"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{query}",
	}
	queriesFailed = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "queries_failed",
		Help:         "The number of channel client queries that failed (timeouts excluded).",
		LabelNames:   []string{"chaincode", "Fcn", "fail"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{query}.%{fail}",
	}
	queryTimeouts = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "query_timeouts",
		Help:         "The number of channel queries that have failed due to time out.",
		LabelNames:   []string{"chaincode", "Fcn", "fail"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{query}.%{timeout}",
	}
	queryDuration = metrics.HistogramOpts{
		Namespace:    "channel",
		Name:         "query_duration",
		Help:         "The time to complete channel client query.",
		LabelNames:   []string{"chaincode", "Fcn"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{query}",
	}
	executionsReceived = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "executions_received",
		Help:         "The number of channel client executions received.",
		LabelNames:   []string{"chaincode", "Fcn"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{execute}",
	}
	executionsFailed = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "executions_failed",
		Help:         "The number of channel client executions that failed (timeouts excluded).",
		LabelNames:   []string{"chaincode", "Fcn", "fail"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{execute}.%{fail}",
	}
	executionTimeouts = metrics.CounterOpts{
		Namespace:    "channel",
		Name:         "execution_timeouts",
		Help:         "The number of channel executions that have failed due to time out.",
		LabelNames:   []string{"chaincode", "Fcn", "fail"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{execute}.%{timeout}",
	}
	executionDuration = metrics.HistogramOpts{
		Namespace:    "channel",
		Name:         "execution_duration",
		Help:         "The time to complete channel client execution.",
		LabelNames:   []string{"chaincode", "Fcn"},
		StatsdFormat: "%{#fqname}.%{type}.%{channel}.%{execution}",
	}
)

// ClientMetrics contains the metrics used in the (channel) client
type ClientMetrics struct {
	QueriesReceived    metrics.Counter
	QueriesFailed      metrics.Counter
	QueryDuration      metrics.Histogram
	QueryTimeouts      metrics.Counter
	ExecutionsReceived metrics.Counter
	ExecutionsFailed   metrics.Counter
	ExecutionDuration  metrics.Histogram
	ExecutionTimeouts  metrics.Counter
}

// NewClientMetrics builds a new instance of ClientMetrics
func NewClientMetrics(p metrics.Provider) *ClientMetrics {
	return &ClientMetrics{
		QueriesReceived:    p.NewCounter(queriesReceived),
		QueriesFailed:      p.NewCounter(queriesFailed),
		QueryDuration:      p.NewHistogram(queryDuration),
		QueryTimeouts:      p.NewCounter(queryTimeouts),
		ExecutionsReceived: p.NewCounter(executionsReceived),
		ExecutionsFailed:   p.NewCounter(executionsFailed),
		ExecutionDuration:  p.NewHistogram(executionDuration),
		ExecutionTimeouts:  p.NewCounter(executionTimeouts),
	}
}
