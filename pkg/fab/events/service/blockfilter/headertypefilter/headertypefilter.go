/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package headertypefilter

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
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
		env, err := protoutil.ExtractEnvelope(block, i)
		if err != nil {
			logger.Errorf("error extracting envelope from block: %s", err)
			continue
		}
		payload, err := protoutil.UnmarshalPayload(env.Payload)
		if err != nil {
			logger.Errorf("error extracting payload from block: %s", err)
			continue
		}
		chdr, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
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
