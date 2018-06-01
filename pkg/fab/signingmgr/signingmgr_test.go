/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package signingmgr

import (
	"bytes"
	"testing"

	bccspwrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

func TestSigningManager(t *testing.T) {

	signingMgr, err := New(&fcmocks.MockCryptoSuite{})
	if err != nil {
		t.Fatalf("Failed to  setup discovery provider: %s", err)
	}

	_, err = signingMgr.Sign(nil, nil)
	if err == nil {
		t.Fatal("Should have failed to sign nil object")
	}

	_, err = signingMgr.Sign([]byte(""), nil)
	if err == nil {
		t.Fatal("Should have failed to sign object empty object")
	}

	_, err = signingMgr.Sign([]byte("Hello"), nil)
	if err == nil {
		t.Fatal("Should have failed to sign object with nil key")
	}

	signedObj, err := signingMgr.Sign([]byte("Hello"), bccspwrapper.GetKey(&mockmsp.MockKey{}))
	if err != nil {
		t.Fatalf("Failed to sign object: %s", err)
	}

	expectedObj := []byte("testSignature")
	if !bytes.Equal(signedObj, expectedObj) {
		t.Fatalf("Expecting %s, got %s", expectedObj, signedObj)
	}

}
