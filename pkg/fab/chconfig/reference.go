/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// Ref channel configuration lazy reference
type Ref struct {
	*lazyref.Reference
	pvdr      Provider
	ctx       fab.ClientContext
	channelID string
}

// NewRef returns a new channel config reference
func NewRef(refresh time.Duration, pvdr Provider, channel string, ctx fab.ClientContext) *Ref {
	cfgRef := &Ref{
		pvdr:      pvdr,
		ctx:       ctx,
		channelID: channel,
	}

	cfgRef.Reference = lazyref.New(
		cfgRef.initializer(),
		lazyref.WithRefreshInterval(lazyref.InitImmediately, refresh),
	)

	return cfgRef
}

func (ref *Ref) initializer() lazyref.Initializer {
	return func() (interface{}, error) {
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
}
