/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockfilter

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// AcceptAny returns a block filter that accepts any block
var AcceptAny fab.BlockFilter = func(block *cb.Block) bool {
	return true
}
