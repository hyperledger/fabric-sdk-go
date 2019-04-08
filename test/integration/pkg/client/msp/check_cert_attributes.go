/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib/attrmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// nolint: deadcode
func checkCertAttributes(t *testing.T, certBytes []byte, expected []msp.Attribute) {
	attrs, err := getCertAttributes(certBytes)
	require.NoError(t, err)
	for _, a := range expected {
		v, ok, err := attrs.Value(a.Name)
		require.NoError(t, err)
		require.True(t, attrs.Contains(a.Name), "does not contain attribute '%s'", a.Name)
		require.True(t, ok, "attribute '%s' was not found", a.Name)
		require.True(t, v == a.Value, "incorrect value for '%s'; expected '%s' but found '%s'", a.Name, a.Value, v)
	}
}

func getCertAttributes(certBytes []byte) (*attrmgr.Attributes, error) {
	decoded, _ := pem.Decode(certBytes)
	if decoded == nil {
		return nil, errors.New("Failed cert decoding")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		return nil, errors.Errorf("failed to parse certificate: %s", err)
	}
	if cert == nil {
		return nil, errors.Errorf("failed to parse certificate: %s", err)
	}
	mgr := attrmgr.New()
	attrs, err := mgr.GetAttributesFromCert(cert)
	if err != nil {
		return nil, errors.Errorf("Failed to GetAttributesFromCert: %s", err)
	}

	return attrs, nil
}
