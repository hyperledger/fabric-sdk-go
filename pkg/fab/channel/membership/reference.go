/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// Ref membership reference that refreshes to load the given channel config reference
type Ref struct {
	*lazyref.Reference
	chConfigRef *lazyref.Reference
	context     Context
	// Note: the following variables are only accessed from Ref.initializer which is synchronized
	configBlockNumber uint64
	mem               fab.ChannelMembership
}

// NewRef returns a new membership reference
func NewRef(refresh time.Duration, context Context, chConfigRef *lazyref.Reference) *Ref {
	ref := &Ref{
		chConfigRef: chConfigRef,
		context:     context,
	}

	ref.Reference = lazyref.New(
		ref.initializer(),
		lazyref.WithRefreshInterval(lazyref.InitImmediately, refresh),
	)

	return ref
}

// Validate calls validate on the underlying reference
func (ref *Ref) Validate(serializedID []byte) error {
	membership, err := ref.get()
	if err != nil {
		return err
	}
	return membership.Validate(serializedID)
}

// Verify calls validate on the underlying reference
func (ref *Ref) Verify(serializedID []byte, msg []byte, sig []byte) error {
	membership, err := ref.get()
	if err != nil {
		return err
	}
	return membership.Verify(serializedID, msg, sig)
}

// ContainsMSP checks if given MSP is available in the underlying reference
func (ref *Ref) ContainsMSP(msp string) bool {
	membership, err := ref.get()
	if err != nil {
		logger.Debugf("Failed to check ContainsMSP, err: %s", err)
		return false
	}
	return membership.ContainsMSP(msp)
}

func (ref *Ref) get() (fab.ChannelMembership, error) {
	m, err := ref.Get()
	if err != nil {
		return nil, err
	}
	return m.(fab.ChannelMembership), nil
}

func (ref *Ref) initializer() lazyref.Initializer {
	return func() (interface{}, error) {
		logger.Debugf("Initializing membership reference...")

		channelCfg, err := ref.chConfigRef.Get()
		if err != nil {
			return nil, errors.WithMessage(err, "could not get channel config from reference")
		}
		cfg, ok := channelCfg.(fab.ChannelCfg)
		if !ok {
			return nil, errors.New("chConfigRef.Get() returned unexpected value ")
		}

		logger.Debugf("Got config block with number %d have %d", cfg.BlockNumber(), ref.configBlockNumber)

		// Membership is refreshed only if we have a newer config block
		if ref.mem == nil || cfg.BlockNumber() > ref.configBlockNumber {
			logger.Debugf("Creating membership for channel [%s]...", cfg.ID())
			ref.mem, err = New(ref.context, cfg)
			if err != nil {
				return nil, err
			}
			ref.configBlockNumber = cfg.BlockNumber()
		}

		return ref.mem, nil
	}
}
