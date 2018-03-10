/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var blockInterests = []*pb.Interest{&pb.Interest{EventType: pb.EventType_BLOCK}}
var filteredBlockInterests = []*pb.Interest{&pb.Interest{EventType: pb.EventType_FILTEREDBLOCK}}

type params struct {
	connProvider      api.ConnectionProvider
	interests         []*pb.Interest
	permitBlockEvents bool
	respTimeout       time.Duration
}

func defaultParams() *params {
	return &params{
		connProvider:      ehConnProvider,
		interests:         blockInterests,
		respTimeout:       5 * time.Second,
		permitBlockEvents: true,
	}
}

// WithBlockEvents indicates that block events are to be received.
func WithBlockEvents() options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderAndInterestsSetter); ok {
			setter.SetConnectionProviderAndInterests(ehConnProvider, blockInterests, true)
		}
	}
}

// withConnectionProviderAndInterests is used only for testing
func withConnectionProviderAndInterests(connProvider api.ConnectionProvider, interests []*pb.Interest, permitBlockEvents bool) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderAndInterestsSetter); ok {
			setter.SetConnectionProviderAndInterests(connProvider, interests, permitBlockEvents)
		}
	}
}

type connectionProviderAndInterestsSetter interface {
	SetConnectionProviderAndInterests(connProvider api.ConnectionProvider, interests []*pb.Interest, permitBlockEvents bool)
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}

func (p *params) SetConnectionProviderAndInterests(connProvider api.ConnectionProvider, interests []*pb.Interest, permitBlockEvents bool) {
	logger.Debugf("ConnProvider: %#v, Interests: %#v, PermitBlockEvents: %t", connProvider, interests, permitBlockEvents)
	p.connProvider = connProvider
	p.interests = interests
	p.permitBlockEvents = permitBlockEvents
}
