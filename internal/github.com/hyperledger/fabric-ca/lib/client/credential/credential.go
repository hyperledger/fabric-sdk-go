/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package credential

import (
	"net/http"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
)

// Credential represents an credential of an identity
type Credential interface {
	// Type returns type of this credential
	Type() string
	// EnrollmentID returns enrollment ID associated with this credential
	// Returns an error if the credential value is not set (SetVal is not called)
	// or not loaded from the disk (Load is not called)
	EnrollmentID() (string, error)
	// Val returns credential value.
	// Returns an error if the credential value is not set (SetVal is not called)
	// or not loaded from the disk (Load is not called)
	Val() (interface{}, error)
	// Sets the credential value
	SetVal(val interface{}) error
	// Stores the credential value to disk
	Store() error
	// Loads the credential value from disk and sets the value of this credential
	Load() error
	// CreateToken returns authorization token for the specified request with
	// specified body
	CreateToken(req *http.Request, reqBody []byte, fabCACompatibilityMode bool) (string, error)
	// Submits revoke request to the Fabric CA server to revoke this credential
	RevokeSelf() (*api.RevocationResponse, error)
}
