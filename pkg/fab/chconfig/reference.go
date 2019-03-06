/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// Ref channel configuration lazy reference
type Ref struct {
	*lazyref.Reference
	pvdr       Provider
	ctx        fab.ClientContext
	channelID  string
	errHandler fab.ErrorHandler
}

// ChannelConfigError is returned when the channel config could not be refreshed
type ChannelConfigError error

// NewRef returns a new channel config reference
func NewRef(ctx fab.ClientContext, pvdr Provider, channel string, opts ...options.Opt) *Ref {
	params := newDefaultParams()
	options.Apply(params, opts)

	cfgRef := &Ref{
		pvdr:       pvdr,
		ctx:        ctx,
		channelID:  channel,
		errHandler: params.errHandler,
	}

	cfgRef.Reference = lazyref.New(
		cfgRef.initializer(),
		lazyref.WithRefreshInterval(lazyref.InitImmediately, params.refreshInterval),
	)

	return cfgRef
}

func (ref *Ref) initializer() lazyref.Initializer {
	return func() (interface{}, error) {
		chConfig, err := ref.getConfig()
		if err != nil && ref.errHandler != nil {
			logger.Debugf("[%s] An error occurred while retrieving channel config. Invoking error handler.", ref.channelID)
			ref.errHandler(ref.ctx, ref.channelID, ChannelConfigError(err))
		}
		return chConfig, err
	}
}

func (ref *Ref) getConfig() (fab.ChannelCfg, error) {
	chConfigProvider, err := ref.pvdr(ref.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating channel config provider")
	}

	reqCtx, cancel := contextImpl.NewRequest(ref.ctx, contextImpl.WithTimeoutType(fab.PeerResponse))
	defer cancel()

	chConfig, err := chConfigProvider.Query(reqCtx)
	if err != nil {
		return nil, err
	}

	return chConfig, nil
}
