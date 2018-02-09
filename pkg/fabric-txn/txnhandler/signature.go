/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

//NewSignatureValidationHandler returns a handler that validates an endorsement
func NewSignatureValidationHandler(next ...chclient.Handler) *SignatureValidationHandler {
	return &SignatureValidationHandler{next: getNext(next)}
}

//SignatureValidationHandler for transaction proposal response filtering
type SignatureValidationHandler struct {
	next chclient.Handler
}

//Handle for Filtering proposal response
func (f *SignatureValidationHandler) Handle(requestContext *chclient.RequestContext, clientContext *chclient.ClientContext) {

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

func (f *SignatureValidationHandler) validate(txProposalResponse []*apifabclient.TransactionProposalResponse, ctx *chclient.ClientContext) error {

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

func verifyProposalResponse(res *pb.ProposalResponse, ctx *chclient.ClientContext) error {

	if res.GetEndorsement() == nil {
		return errors.Errorf("Missing endorsement in proposal response")
	}

	serializedIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(res.GetEndorsement().Endorser, serializedIdentity); err != nil {
		return errors.WithMessage(err, "Unmarshal endorser error")
	}

	if ctx.Channel.MSPManager() == nil {
		return errors.Errorf("Channel %s msp manager is nil", ctx.Channel.Name())
	}

	msps, err := ctx.Channel.MSPManager().GetMSPs()
	if err != nil {
		return errors.WithMessage(err, "GetMSPs return error:%v")
	}
	if len(msps) == 0 {
		return errors.Errorf("Channel %s msps is empty", ctx.Channel.Name())
	}

	msp := msps[serializedIdentity.Mspid]
	if msp == nil {
		return errors.Errorf("MSP %s not found", serializedIdentity.Mspid)
	}

	creator, err := msp.DeserializeIdentity(res.GetEndorsement().Endorser)
	if err != nil {
		return errors.WithMessage(err, "Failed to deserialize creator identity")
	}

	// ensure that creator is a valid certificate
	err = creator.Validate()
	if err != nil {
		return errors.WithMessage(err, "The creator certificate is not valid")
	}

	// check the signature against the endorser and payload hash
	digest := append(res.GetPayload(), res.GetEndorsement().Endorser...)

	// validate the signature
	err = creator.Verify(digest, res.GetEndorsement().Signature)
	if err != nil {
		return errors.WithMessage(err, "The creator's signature over the proposal is not valid")
	}

	return nil
}
