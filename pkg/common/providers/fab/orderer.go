/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"

	"github.com/hyperledger/fabric-protos-go/common"
)

// Orderer The Orderer class represents a peer in the target blockchain network to which
// HFC sends a block of transactions of endorsed proposals requiring ordering.
type Orderer interface {
	URL() string
	SendBroadcast(ctx reqContext.Context, envelope *SignedEnvelope) (*common.Status, error)
	SendDeliver(ctx reqContext.Context, envelope *SignedEnvelope) (chan *common.Block, chan error)
}

// A SignedEnvelope can can be sent to an orderer for broadcasting
type SignedEnvelope struct {
	Payload   []byte
	Signature []byte
}
