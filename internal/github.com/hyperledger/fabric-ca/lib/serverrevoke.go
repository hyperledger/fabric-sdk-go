/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/api"
)

type revocationResponseNet struct {
	RevokedCerts []api.RevokedCert
	CRL          string
}

// CertificateStatus represents status of an enrollment certificate
type CertificateStatus string

const (
	// Revoked is the status of a revoked certificate
	Revoked CertificateStatus = "revoked"
	// Good is the status of a active certificate
	Good = "good"
)
