/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package headertypefilter

import (
	"testing"

	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func TestHeaderTypeBlockFilter(t *testing.T) {
	filter := New(cb.HeaderType_CONFIG, cb.HeaderType_CONFIG_UPDATE)

	if !filter(servicemocks.NewBlock("somechannel", servicemocks.NewTransaction("txid", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG))) {
		t.Fatalf("expecting block filter to accept block with header type %s", cb.HeaderType_CONFIG)
	}
	if !filter(servicemocks.NewBlock("somechannel", servicemocks.NewTransaction("txid", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE))) {
		t.Fatalf("expecting block filter to accept block with header type %s", cb.HeaderType_CONFIG_UPDATE)
	}
	if filter(servicemocks.NewBlock("somechannel", servicemocks.NewTransaction("txid", pb.TxValidationCode_VALID, cb.HeaderType_MESSAGE))) {
		t.Fatalf("expecting block filter to reject block with header type %s", cb.HeaderType_MESSAGE)
	}
}
