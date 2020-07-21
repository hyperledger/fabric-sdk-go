/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package api

const (
	// IdemixTokenVersion1 represents version 1 of the authorization token created using Idemix credential
	IdemixTokenVersion1 = "1"
)

// CAInfoResponseNet is the response to the GET /info request
type CAInfoResponseNet struct {
	// CAName is a unique name associated with fabric-ca-server's CA
	CAName string
	// Base64 encoding of PEM-encoded certificate chain
	CAChain string
	// Base64 encoding of Idemix issuer public key
	IssuerPublicKey string
	// Base64 encoding of PEM-encoded Idemix issuer revocation public key
	IssuerRevocationPublicKey string
	// Version of the server
	Version string
}

// EnrollmentResponseNet is the response to the /enroll request
type EnrollmentResponseNet struct {
	// Base64 encoded PEM-encoded ECert
	Cert string
	// The server information
	ServerInfo CAInfoResponseNet
}

// IdemixEnrollmentResponseNet is the response to the /idemix/credential request
type IdemixEnrollmentResponseNet struct {
	// Base64 encoding of proto bytes of idemix.Credential
	Credential string
	// Attribute name-value pairs
	Attrs map[string]interface{}
	// Base64 encoding of proto bytes of idemix.CredentialRevocationInformation
	CRI string
	// Base64 encoding of the issuer nonce
	Nonce string
	// The CA information
	CAInfo CAInfoResponseNet
}
