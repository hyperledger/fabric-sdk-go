/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metrics

// TODO remove this package once the Fabric copy is imported

import (
	"errors"
	"net/http"
	"time"

	"net"

	"sort"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
	statsdreporter "github.com/uber-go/tally/statsd"
)

var logger = logging.NewLogger("perf_tally_provider")

func newRootScope(opts tally.ScopeOptions, interval time.Duration) tally.Scope {
	s, _ := tally.NewRootScope(opts, interval)
	return s
}

func newStatsdReporter(statsdReporterOpts StatsdReporterOpts) (tally.StatsReporter, error) {
	if statsdReporterOpts.Address == "" {
		return nil, errors.New("missing statsd server Address option")
	}

	if statsdReporterOpts.FlushInterval <= 0 {
		return nil, errors.New("missing statsd FlushInterval option")
	}

	if statsdReporterOpts.FlushBytes <= 0 {
		return nil, errors.New("missing statsd FlushBytes option")
	}

	statter, err := statsd.NewBufferedClient(statsdReporterOpts.Address,
		"", statsdReporterOpts.FlushInterval, statsdReporterOpts.FlushBytes)
	if err != nil {
		return nil, err
	}
	opts := statsdreporter.Options{}
	reporter := statsdreporter.NewReporter(statter, opts)
	statsdReporter := &statsdReporter{StatsReporter: reporter, statter: statter}
	return statsdReporter, nil
}

func newPromReporter(promReporterOpts PromReporterOpts) (promreporter.Reporter, error) {
	if promReporterOpts.ListenAddress == "" {
		return nil, errors.New("missing prometheus listenAddress option")
	}

	opts := promreporter.Options{Registerer: prometheus.NewRegistry()}
	reporter := promreporter.NewReporter(opts)
	mux := http.NewServeMux()
	handler := promReporterHTTPHandler(opts.Registerer.(*prometheus.Registry))
	mux.Handle("/metrics", handler)
	server := &http.Server{Handler: mux}
	addr := promReporterOpts.ListenAddress
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	promReporter := &promReporter{
		Reporter: reporter,
		server:   server,
		registry: opts.Registerer.(*prometheus.Registry),
		listener: listener}
	go func() {
		err := server.Serve(listener)
		if err != nil {
			logger.Warnf("failed to start registry server: %s", err)
		}
	}()
	return promReporter, nil
}

type statsdReporter struct {
	tally.StatsReporter
	statter statsd.Statter
}

type promReporter struct {
	promreporter.Reporter
	server   *http.Server
	listener net.Listener
	registry *prometheus.Registry
}

func (r *statsdReporter) Close() error {
	return r.statter.Close()
}

func (r *statsdReporter) ReportCounter(name string, tags map[string]string, value int64) {
	r.StatsReporter.ReportCounter(tagsToName(name, tags), tags, value)
}

func (r *statsdReporter) ReportGauge(name string, tags map[string]string, value float64) {
	r.StatsReporter.ReportGauge(tagsToName(name, tags), tags, value)
}

func (r *statsdReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	r.StatsReporter.ReportTimer(tagsToName(name, tags), tags, interval)
}

func (r *statsdReporter) ReportHistogramValueSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	r.StatsReporter.ReportHistogramValueSamples(tagsToName(name, tags), tags, buckets, bucketLowerBound, bucketUpperBound, samples)
}

func (r *statsdReporter) ReportHistogramDurationSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	r.StatsReporter.ReportHistogramDurationSamples(tagsToName(name, tags), tags, buckets, bucketLowerBound, bucketUpperBound, samples)
}

func (r *statsdReporter) Capabilities() tally.Capabilities {
	return r
}

func (r *statsdReporter) Reporting() bool {
	return true
}

func (r *statsdReporter) Tagging() bool {
	return true
}

func (r *promReporter) Close() error {
	//TODO: Shutdown server gracefully?
	// Close() is not a graceful way since it closes server immediately
	err := r.server.Close()
	if err != nil {
		logger.Warnf("failed to close registry server: %s", err)
	}
	err = r.listener.Close()
	if err != nil {
		logger.Warnf("failed to close registrey listener: %s", err)
	}
	return err
}

func (r *promReporter) HTTPHandler() http.Handler {
	return promReporterHTTPHandler(r.registry)
}

func promReporterHTTPHandler(registry prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func tagsToName(name string, tags map[string]string) string {
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		name = name + promreporter.DefaultSeparator + k + "-" + tags[k]
	}

	return name
}
