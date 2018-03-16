/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package headertypefilter

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.NewLogger("eventservice/blockfilter")

// New returns a block filter that filters out blocks that
// don't contain envelopes of the given type(s)
func New(headerTypes ...cb.HeaderType) fab.BlockFilter {
	return func(block *cb.Block) bool {
		return hasType(block, headerTypes...)
	}
}

func hasType(block *cb.Block, headerTypes ...cb.HeaderType) bool {
	for i := 0; i < len(block.Data.Data); i++ {
		env, err := utils.ExtractEnvelope(block, i)
		if err != nil {
			logger.Errorf("error extracting envelope from block: %s", err)
			continue
		}
		payload, err := utils.ExtractPayload(env)
		if err != nil {
			logger.Errorf("error extracting payload from block: %s", err)
			continue
		}
		chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			logger.Errorf("error extracting channel header: %s", err)
			continue
		}
		htype := cb.HeaderType(chdr.Type)
		for _, headerType := range headerTypes {
			if htype == headerType {
				return true
			}
		}
	}
	return false
}
