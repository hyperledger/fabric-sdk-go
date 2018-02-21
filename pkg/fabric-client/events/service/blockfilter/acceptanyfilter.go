/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockfilter

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// AcceptAny returns a block filter that accepts any block
var AcceptAny apifabclient.BlockFilter = func(block *cb.Block) bool {
	return true
}
