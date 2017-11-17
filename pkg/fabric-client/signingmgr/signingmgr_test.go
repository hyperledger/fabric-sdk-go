/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package signingmgr

import (
	"bytes"
	"testing"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client/mocks"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestSigningManager(t *testing.T) {

	signingMgr, err := NewSigningManager(&fcmocks.MockCryptoSuite{}, &fcmocks.MockConfig{})
	if err != nil {
		t.Fatalf("Failed to  setup discovery provider: %s", err)
	}

	_, err = signingMgr.Sign(nil, nil)
	if err == nil {
		t.Fatalf("Should have failed to sign nil object")
	}

	_, err = signingMgr.Sign([]byte(""), nil)
	if err == nil {
		t.Fatalf("Should have failed to sign object empty object")
	}

	_, err = signingMgr.Sign([]byte("Hello"), nil)
	if err == nil {
		t.Fatalf("Should have failed to sign object with nil key")
	}

	signedObj, err := signingMgr.Sign([]byte("Hello"), cryptosuite.GetKey(&mocks.MockKey{}))
	if err != nil {
		t.Fatalf("Failed to sign object: %s", err)
	}

	expectedObj := []byte("testSignature")
	if !bytes.Equal(signedObj, expectedObj) {
		t.Fatalf("Expecting %s, got %s", expectedObj, signedObj)
	}

}
