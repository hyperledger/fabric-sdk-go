/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package crypto

import (
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// LocalSigner is a temporary stub interface which will be implemented by the local MSP
type LocalSigner interface {
	SignatureHeaderMaker
	Signer
}

// Signer signs messages
type Signer interface {
	// Sign a message and return the signature over the digest, or error on failure
	Sign(message []byte) ([]byte, error)
}

// IdentitySerializer serializes identities
type IdentitySerializer interface {
	// Serialize converts an identity to bytes
	Serialize() ([]byte, error)
}

// SignatureHeaderMaker creates a new SignatureHeader
type SignatureHeaderMaker interface {
	// NewSignatureHeader creates a SignatureHeader with the correct signing identity and a valid nonce
	NewSignatureHeader() (*cb.SignatureHeader, error)
}

// SignatureHeaderCreator creates signature headers
type SignatureHeaderCreator struct {
	SignerSupport
}

// SignerSupport implements the needed support for LocalSigner
type SignerSupport interface {
	Signer
	IdentitySerializer
}
