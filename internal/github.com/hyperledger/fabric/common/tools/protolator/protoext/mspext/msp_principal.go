/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package mspext

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/msp"
)

type MSPPrincipal struct{ *msp.MSPPrincipal }

func (mp *MSPPrincipal) Underlying() proto.Message {
	return mp.MSPPrincipal
}

func (mp *MSPPrincipal) VariablyOpaqueFields() []string {
	return []string{"principal"}
}

func (mp *MSPPrincipal) VariablyOpaqueFieldProto(name string) (proto.Message, error) {
	if name != mp.VariablyOpaqueFields()[0] {
		return nil, fmt.Errorf("not a marshaled field: %s", name)
	}
	switch mp.PrincipalClassification {
	case msp.MSPPrincipal_ROLE:
		return &msp.MSPRole{}, nil
	case msp.MSPPrincipal_ORGANIZATION_UNIT:
		return &msp.OrganizationUnit{}, nil
	case msp.MSPPrincipal_IDENTITY:
		return &msp.SerializedIdentity{}, nil
	default:
		return nil, fmt.Errorf("unable to decode MSP type: %v", mp.PrincipalClassification)
	}
}
