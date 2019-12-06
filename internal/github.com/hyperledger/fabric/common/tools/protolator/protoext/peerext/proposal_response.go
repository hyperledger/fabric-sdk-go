/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package peerext

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type ProposalResponsePayload struct {
	*peer.ProposalResponsePayload
}

func (ppr *ProposalResponsePayload) Underlying() proto.Message {
	return ppr.ProposalResponsePayload
}

func (ppr *ProposalResponsePayload) StaticallyOpaqueFields() []string {
	return []string{"extension"}
}

func (ppr *ProposalResponsePayload) StaticallyOpaqueFieldProto(name string) (proto.Message, error) {
	if name != ppr.StaticallyOpaqueFields()[0] {
		return nil, fmt.Errorf("not a marshaled field: %s", name)
	}
	return &peer.ChaincodeAction{}, nil
}
