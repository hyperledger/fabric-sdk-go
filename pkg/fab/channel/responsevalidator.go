/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// TransactionProposalResponseVerifier struct is for verifying TransactionProposalResponse and matches config blocks
type TransactionProposalResponseVerifier struct {
	MinResponses int
}

// Verify checks transaction proposal response (empty)
func (tprv *TransactionProposalResponseVerifier) Verify(response *fab.TransactionProposalResponse) error {
	return nil
}

// Match verifies and matches transaction proposal responses
func (tprv *TransactionProposalResponseVerifier) Match(transactionProposalResponses []*fab.TransactionProposalResponse) error {
	if tprv.MinResponses <= 0 {
		return errors.New("minimum Responses has to be greater than zero")
	}

	if len(transactionProposalResponses) < tprv.MinResponses {
		return errors.Errorf("required minimum %d endorsments got %d", tprv.MinResponses, len(transactionProposalResponses))
	}

	block, err := createCommonBlock(transactionProposalResponses[0])
	if err != nil {
		return err
	}

	if block.Data == nil || block.Data.Data == nil {
		return errors.New("config block data is nil")
	}

	if len(block.Data.Data) != 1 {
		return errors.New("config block must contain one transaction")
	}

	// Compare block data from  remaining responses
	for _, tpr := range transactionProposalResponses[1:] {
		b, err := createCommonBlock(tpr)
		if err != nil {
			return err
		}

		if !proto.Equal(block.Data, b.Data) {
			return errors.WithStack(status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(), "payloads for config block do not match", nil))
		}
	}

	return nil
}
