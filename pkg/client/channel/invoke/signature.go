/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/status"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

//NewSignatureValidationHandler returns a handler that validates an endorsement
func NewSignatureValidationHandler(next ...Handler) *SignatureValidationHandler {
	return &SignatureValidationHandler{next: getNext(next)}
}

//SignatureValidationHandler for transaction proposal response filtering
type SignatureValidationHandler struct {
	next Handler
}

//Handle for Filtering proposal response
func (f *SignatureValidationHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {
	//Filter tx proposal responses
	err := f.validate(requestContext.Response.Responses, clientContext)
	if err != nil {
		requestContext.Error = errors.WithMessage(err, "endorsement validation failed")
		return
	}

	// Delegate to next step if any
	if f.next != nil {
		f.next.Handle(requestContext, clientContext)
	}
}

func (f *SignatureValidationHandler) validate(txProposalResponse []*fab.TransactionProposalResponse, ctx *ClientContext) error {
	for _, r := range txProposalResponse {
		if r.ProposalResponse.GetResponse().Status != int32(common.Status_SUCCESS) {
			return status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
		}

		if err := verifyProposalResponse(r.ProposalResponse, ctx); err != nil {
			return err
		}
	}

	return nil
}

func verifyProposalResponse(res *pb.ProposalResponse, ctx *ClientContext) error {
	if res.GetEndorsement() == nil {
		return errors.Errorf("Missing endorsement in proposal response")
	}
	creatorID := res.GetEndorsement().Endorser

	err := ctx.Membership.Validate(creatorID)
	if err != nil {
		return errors.WithMessage(err, "The creator certificate is not valid")
	}

	// check the signature against the endorser and payload hash
	digest := append(res.GetPayload(), res.GetEndorsement().Endorser...)

	// validate the signature
	err = ctx.Membership.Verify(creatorID, digest, res.GetEndorsement().Signature)
	if err != nil {
		return errors.WithMessage(err, "The creator's signature over the proposal is not valid")
	}

	return nil
}
