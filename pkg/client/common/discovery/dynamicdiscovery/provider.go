/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// Provider implements a dynamic Discovery Provider that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type Provider struct {
	cache *lazycache.Cache
}

// Opt is a provider option
type Opt func(o *options)

// WithRefreshInterval sets the interval in which the
// peer cache is refreshed
func WithRefreshInterval(value time.Duration) Opt {
	return func(o *options) {
		o.refreshInterval = value
	}
}

// WithResponseTimeout sets the Discover service response timeout
func WithResponseTimeout(value time.Duration) Opt {
	return func(o *options) {
		o.responseTimeout = value
	}
}

type options struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
}

// New creates a new dynamic discovery provider
func New(config fab.EndpointConfig, opts ...Opt) *Provider {
	options := options{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.refreshInterval == 0 {
		options.refreshInterval = config.Timeout(fab.DiscoveryServiceRefresh)
	}
	if options.responseTimeout == 0 {
		options.responseTimeout = config.Timeout(fab.DiscoveryResponse)
	}

	return &Provider{
		cache: lazycache.New("Dynamic_Discovery_Service_Cache", func(key lazycache.Key) (interface{}, error) {
			if key.String() == "" {
				return newLocalService(options), nil
			}
			return newChannelService(options), nil
		}),
	}
}

// CreateDiscoveryService returns a discovery service for the given channel
func (p *Provider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	ref, err := p.cache.Get(lazycache.NewStringKey(channelID))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get discovery service from cache")
	}
	return ref.(fab.DiscoveryService), nil
}

// CreateLocalDiscoveryService returns a local discovery service
func (p *Provider) CreateLocalDiscoveryService() (fab.DiscoveryService, error) {
	ref, err := p.cache.Get(lazycache.NewStringKey(""))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get local discovery service from cache")
	}
	return ref.(fab.DiscoveryService), nil
}

// Close will close the cache and all services contained by the cache.
func (p *Provider) Close() {
	p.cache.Close()
}
