/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// EventEndpoint extends a Peer endpoint and provides the
// event URL, which may or may not be the same as the Peer URL
type EventEndpoint interface {
	fab.Peer

	// Opts returns additional options for the connection
	Opts() []options.Opt
}
