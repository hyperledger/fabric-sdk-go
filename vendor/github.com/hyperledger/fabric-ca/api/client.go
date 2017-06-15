/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/hyperledger/fabric-ca/lib/tcert"
)

// RegistrationRequest for a new identity
type RegistrationRequest struct {
	// Name is the unique name of the identity
	Name string `json:"id" help:"Unique name of the identity"`
	// Type of identity being registered (e.g. "peer, app, user")
	Type string `json:"type" help:"Type of identity being registered (e.g. 'peer, app, user')"`
	// Secret is an optional password.  If not specified,
	// a random secret is generated.  In both cases, the secret
	// is returned in the RegistrationResponse.
	Secret string `json:"secret,omitempty" help:"The enrollment secret for the identity being registered"`
	// MaxEnrollments is the maximum number of times the secret can
	// be reused to enroll.
	MaxEnrollments int `json:"max_enrollments,omitempty" def:"-1" help:"The maximum number of times the secret can be reused to enroll."`
	// is returned in the response.
	// The identity's affiliation.
	// For example, an affiliation of "org1.department1" associates the identity with "department1" in "org1".
	Affiliation string `json:"affiliation" help:"The identity's affiliation"`
	// Attributes associated with this identity
	Attributes []Attribute `json:"attrs,omitempty"`
	// CAName is the name of the CA to connect to
	CAName string `json:"caname,omitempty" skip:"true"`
}

// RegistrationResponse is a registration response
type RegistrationResponse struct {
	// The secret returned from a successful registration response
	Secret string `json:"secret"`
}

// EnrollmentRequest is a request to enroll an identity
type EnrollmentRequest struct {
	// The identity name to enroll
	Name string `json:"name" skip:"true"`
	// The secret returned via Register
	Secret string `json:"secret,omitempty" skip:"true"`
	// Profile is the name of the signing profile to use in issuing the certificate
	Profile string `json:"profile,omitempty" help:"Name of the signing profile to use in issuing the certificate"`
	// Label is the label to use in HSM operations
	Label string `json:"label,omitempty" help:"Label to use in HSM operations"`
	// CSR is Certificate Signing Request info
	CSR *CSRInfo `json:"csr,omitempty" help:"Certificate Signing Request info"`
	// CAName is the name of the CA to connect to
	CAName string `json:"caname,omitempty" skip:"true"`
}

// ReenrollmentRequest is a request to reenroll an identity.
// This is useful to renew a certificate before it has expired.
type ReenrollmentRequest struct {
	// Profile is the name of the signing profile to use in issuing the certificate
	Profile string `json:"profile,omitempty"`
	// Label is the label to use in HSM operations
	Label string `json:"label,omitempty"`
	// CSR is Certificate Signing Request info
	CSR *CSRInfo `json:"csr,omitempty"`
	// CAName is the name of the CA to connect to
	CAName string `json:"caname,omitempty" skip:"true"`
}

// RevocationRequest is a revocation request for a single certificate or all certificates
// associated with an identity.
// To revoke a single certificate, both the Serial and AKI fields must be set;
// otherwise, to revoke all certificates and the identity associated with an enrollment ID,
// the Name field must be set to an existing enrollment ID.
// A RevocationRequest can only be performed by a user with the "hf.Revoker" attribute.
type RevocationRequest struct {
	// Name of the identity whose certificates should be revoked
	// If this field is omitted, then Serial and AKI must be specified.
	Name string `json:"id,omitempty" opt:"e" help:"Identity whose certificates should be revoked"`
	// Serial number of the certificate to be revoked
	// If this is omitted, then Name must be specified
	Serial string `json:"serial,omitempty" opt:"s" help:"Serial number of the certificate to be revoked"`
	// AKI (Authority Key Identifier) of the certificate to be revoked
	AKI string `json:"aki,omitempty" opt:"a" help:"AKI (Authority Key Identifier) of the certificate to be revoked"`
	// Reason is the reason for revocation.  See https://godoc.org/golang.org/x/crypto/ocsp for
	// valid values.  The default value is 0 (ocsp.Unspecified).
	Reason string `json:"reason,omitempty" opt:"r" help:"Reason for revocation"`
	// CAName is the name of the CA to connect to
	CAName string `json:"caname,omitempty" skip:"true"`
}

// GetTCertBatchRequest is input provided to identity.GetTCertBatch
type GetTCertBatchRequest struct {
	// Number of TCerts in the batch.
	Count int `json:"count"`
	// The attribute names whose names and values are to be sealed in the issued TCerts.
	AttrNames []string `json:"attr_names,omitempty"`
	// EncryptAttrs denotes whether to encrypt attribute values or not.
	// When set to true, each issued TCert in the batch will contain encrypted attribute values.
	EncryptAttrs bool `json:"encrypt_attrs,omitempty"`
	// Certificate Validity Period.  If specified, the value used
	// is the minimum of this value and the configured validity period
	// of the TCert manager.
	ValidityPeriod time.Duration `json:"validity_period,omitempty"`
	// The pre-key to be used for key derivation.
	PreKey string `json:"prekey"`
	// DisableKeyDerivation if true disables key derivation so that a TCert is not
	// cryptographically related to an ECert.  This may be necessary when using an
	// HSM which does not support the TCert's key derivation function.
	DisableKeyDerivation bool `json:"disable_kdf,omitempty"`
	// CAName is the name of the CA to connect to
	CAName string `json:"caname,omitempty" skip:"true"`
}

// GetTCertBatchResponse is the return value of identity.GetTCertBatch
type GetTCertBatchResponse struct {
	tcert.GetBatchResponse
}

// GetCAInfoRequest is request to get generic CA information
type GetCAInfoRequest struct {
	CAName string `json:"caname,omitempty" skip:"true"`
}

// CSRInfo is Certificate Signing Request information
type CSRInfo struct {
	CN           string               `json:"CN"`
	Names        []csr.Name           `json:"names,omitempty"`
	Hosts        []string             `json:"hosts,omitempty"`
	KeyRequest   *csr.BasicKeyRequest `json:"key,omitempty"`
	CA           *csr.CAConfig        `json:"ca,omitempty"`
	SerialNumber string               `json:"serial_number,omitempty"`
}

// Attribute is a name and value pair
type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
