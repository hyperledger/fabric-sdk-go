/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabca

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// FabricCAClient is the client interface for fabric-ca
type FabricCAClient interface {
	CAName() string
	Enroll(enrollmentID string, enrollmentSecret string) (apicryptosuite.Key, []byte, error)
	// Reenroll to renew user's enrollment certificate
	Reenroll(user User) (apicryptosuite.Key, []byte, error)
	Register(registrar User, request *RegistrationRequest) (string, error)
	Revoke(registrar User, request *RevocationRequest) error
}

// RegistrationRequest defines the attributes required to register a user with the CA
type RegistrationRequest struct {
	// Name is the unique name of the identity
	Name string
	// Type of identity being registered (e.g. "peer, app, user")
	Type string
	// MaxEnrollments is the number of times the secret can  be reused to enroll.
	// if omitted, this defaults to max_enrollments configured on the server
	MaxEnrollments int
	// The identity's affiliation e.g. org1.department1
	Affiliation string
	// Optional attributes associated with this identity
	Attributes []Attribute
	// CAName is the name of the CA to connect to
	CAName string
	// Secret is an optional password.  If not specified,
	// a random secret is generated.  In both cases, the secret
	// is returned from registration.
	Secret string
}

// RevocationRequest defines the attributes required to revoke credentials with the CA
type RevocationRequest struct {
	// Name of the identity whose certificates should be revoked
	// If this field is omitted, then Serial and AKI must be specified.
	Name string
	// Serial number of the certificate to be revoked
	// If this is omitted, then Name must be specified
	Serial string
	// AKI (Authority Key Identifier) of the certificate to be revoked
	AKI string
	// Reason is the reason for revocation. See https://godoc.org/golang.org/x/crypto/ocsp
	// for valid values. The default value is 0 (ocsp.Unspecified).
	Reason string
	// CAName is the name of the CA to connect to
	CAName string
}

// Attribute defines additional attributes that may be passed along during registration
type Attribute struct {
	Key   string
	Value string
}
