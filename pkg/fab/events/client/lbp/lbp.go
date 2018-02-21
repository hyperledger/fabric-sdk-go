/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lbp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// LoadBalancePolicy chooses a peer from a set of peers
type LoadBalancePolicy interface {
	Choose(peers []fab.Peer) (fab.Peer, error)
}
