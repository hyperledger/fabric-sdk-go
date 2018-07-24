/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metrics

// TODO remove this package once the Fabric copy is imported

import (
	"fmt"
	"io"
	"time"

	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
)

const (
	namespace string = "hyperledger_fabric"

	statsdReporterType = "statsd"
	promReporterType   = "prom"

	defaultReporterType = statsdReporterType
	defaultInterval     = 1 * time.Second

	defaultStatsdReporterFlushInterval = 2 * time.Second
	defaultStatsdReporterFlushBytes    = 1432
)

// RootScope tally.NoopScope is a scope that does nothing
var RootScope = tally.NoopScope
var rootScopeMutex = &sync.Mutex{}
var running bool

// StatsdReporterOpts ...
type StatsdReporterOpts struct {
	Address       string
	FlushInterval time.Duration
	FlushBytes    int
}

// PromReporterOpts ...
type PromReporterOpts struct {
	ListenAddress string
}

// Opts ...
type Opts struct {
	Reporter           string
	Interval           time.Duration
	Enabled            bool
	StatsdReporterOpts StatsdReporterOpts
	PromReporterOpts   PromReporterOpts
}

// NewOpts create metrics options based config file.
// TODO: Currently this is only for peer node which uses global viper.
// As for orderer, which uses its local viper, we are unable to get
// metrics options with the function NewOpts()
func NewOpts(peerConfig *viper.Viper) Opts {
	opts := Opts{}
	opts.Enabled = peerConfig.GetBool("metrics.enabled")
	if report := peerConfig.GetString("metrics.reporter"); report != "" {
		opts.Reporter = report
	} else {
		opts.Reporter = defaultReporterType
	}
	if interval := peerConfig.GetDuration("metrics.interval"); interval > 0 {
		opts.Interval = interval
	} else {
		opts.Interval = defaultInterval
	}

	if opts.Reporter == statsdReporterType {
		statsdOpts := StatsdReporterOpts{}
		statsdOpts.Address = peerConfig.GetString("metrics.statsdReporter.address")
		if flushInterval := peerConfig.GetDuration("metrics.statsdReporter.flushInterval"); flushInterval > 0 {
			statsdOpts.FlushInterval = flushInterval
		} else {
			statsdOpts.FlushInterval = defaultStatsdReporterFlushInterval
		}
		if flushBytes := peerConfig.GetInt("metrics.statsdReporter.flushBytes"); flushBytes > 0 {
			statsdOpts.FlushBytes = flushBytes
		} else {
			statsdOpts.FlushBytes = defaultStatsdReporterFlushBytes
		}
		opts.StatsdReporterOpts = statsdOpts
	}

	if opts.Reporter == promReporterType {
		promOpts := PromReporterOpts{}
		promOpts.ListenAddress = peerConfig.GetString("metrics.promReporter.listenAddress")
		opts.PromReporterOpts = promOpts
	}

	return opts
}

// Start starts metrics server
func Start(opts Opts) error {
	if !opts.Enabled {
		return errors.New("Unable to start metrics server because it is disabled")
	}
	rootScopeMutex.Lock()
	defer rootScopeMutex.Unlock()
	if !running {
		rootScope, err := create(opts)
		if err == nil {
			running = true
			RootScope = rootScope
		}
		return err
	}
	return errors.New("metrics server was already started")
}

// Shutdown closes underlying resources used by metrics server
func Shutdown() error {
	rootScopeMutex.Lock()
	defer rootScopeMutex.Unlock()
	if running {
		var err error
		if closer, ok := RootScope.(io.Closer); ok {
			if err = closer.Close(); err != nil {
				return err
			}
		}
		running = false
		RootScope = tally.NoopScope
		return err
	}
	return nil
}

func create(opts Opts) (rootScope tally.Scope, e error) {
	if !opts.Enabled {
		rootScope = tally.NoopScope
	} else {
		if opts.Interval <= 0 {
			e = fmt.Errorf("invalid Interval option %d", opts.Interval)
			return
		}
		var reporter tally.StatsReporter
		var cachedReporter tally.CachedStatsReporter
		switch opts.Reporter {
		case statsdReporterType:
			reporter, e = newStatsdReporter(opts.StatsdReporterOpts)
		case promReporterType:
			cachedReporter, e = newPromReporter(opts.PromReporterOpts)
		default:
			e = fmt.Errorf("not supported Reporter type %s", opts.Reporter)
			return
		}
		if e != nil {
			return
		}
		rootScope = newRootScope(
			tally.ScopeOptions{
				Prefix:         namespace,
				Reporter:       reporter,
				CachedReporter: cachedReporter,
				Separator:      promreporter.DefaultSeparator,
			}, opts.Interval)
	}
	return
}
