// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"crypto/x509"
	"encoding/pem"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/attrmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/stretchr/testify/assert"
)

func checkCertAttributes(t *testing.T, certBytes []byte, expected []msp.Attribute) {
	decoded, _ := pem.Decode(certBytes)
	if decoded == nil {
		t.Fatalf("Failed cert decoding")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	if cert == nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	mgr := attrmgr.New()
	attrs, err := mgr.GetAttributesFromCert(cert)
	if err != nil {
		t.Fatalf("Failed to GetAttributesFromCert: %s", err)
	}
	for _, a := range expected {
		v, ok, err := attrs.Value(a.Name)
		assert.NoError(t, err)
		assert.True(t, attrs.Contains(a.Name), "does not contain attribute '%s'", a.Name)
		assert.True(t, ok, "attribute '%s' was not found", a.Name)
		assert.True(t, v == a.Value, "incorrect value for '%s'; expected '%s' but found '%s'", a.Name, a.Value, v)
	}
}
