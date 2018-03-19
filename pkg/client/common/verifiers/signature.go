/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package verifiers provides various verifiers (e.g. signature)
package verifiers

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/status"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// Signature verifies response signature
type Signature struct {
	Membership fab.ChannelMembership
}

// Verify checks transaction proposal response
func (v *Signature) Verify(response *fab.TransactionProposalResponse) error {

	if response.ProposalResponse.GetResponse().Status != int32(common.Status_SUCCESS) {
		return status.NewFromProposalResponse(response.ProposalResponse, response.Endorser)
	}

	res := response.ProposalResponse

	if res.GetEndorsement() == nil {
		return errors.Errorf("Missing endorsement in proposal response")
	}
	creatorID := res.GetEndorsement().Endorser

	err := v.Membership.Validate(creatorID)
	if err != nil {
		return errors.WithMessage(err, "The creator certificate is not valid")
	}

	// check the signature against the endorser and payload hash
	digest := append(res.GetPayload(), res.GetEndorsement().Endorser...)

	// validate the signature
	err = v.Membership.Verify(creatorID, digest, res.GetEndorsement().Signature)
	if err != nil {
		return errors.WithMessage(err, "The creator's signature over the proposal is not valid")
	}

	return nil
}

// Match matches transaction proposal responses (empty for signature verifier)
func (v *Signature) Match(response []*fab.TransactionProposalResponse) error {
	return nil
}
