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
	connProvider api.ConnectionProvider
	interests    []*pb.Interest
	respTimeout  time.Duration
}

func defaultParams() *params {
	return &params{
		connProvider: ehConnProvider,
		interests:    filteredBlockInterests,
		respTimeout:  5 * time.Second,
	}
}

// withConnectionProvider is used only for testing
func withConnectionProvider(connProvider api.ConnectionProvider) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderSetter); ok {
			setter.SetConnectionProvider(connProvider)
		}
	}
}

type connectionProviderSetter interface {
	SetConnectionProvider(connProvider api.ConnectionProvider)
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}

// SetConnectionProvider is used only for testing
func (p *params) SetConnectionProvider(connProvider api.ConnectionProvider) {
	logger.Debugf("ConnProvider: %#v", connProvider)
	p.connProvider = connProvider
}

func (p *params) PermitBlockEvents() {
	logger.Debugf("PermitBlockEvents")
	p.interests = blockInterests
}
