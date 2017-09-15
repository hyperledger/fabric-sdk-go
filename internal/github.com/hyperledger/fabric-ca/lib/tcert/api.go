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

package tcert

import (
	"math/big"
	"time"
)

/*
 * This file contains definitions of the input and output to the TCert
 * library APIs.
 */

// GetBatchRequest defines input to the GetBatch API
type GetBatchRequest struct {
	// Number of TCerts in the batch.
	Count int `json:"count"`
	// If PublicKeys is non nil, generates a TCert for each public key;
	// in this case, the 'Count' field is ignored and the number of TCerts
	// generated matches the number of public keys in the array.
	PublicKeys [][]byte `json:"public_keys,omitempty"`
	// The attribute name and values that are to be inserted in the issued TCerts.
	Attrs []Attribute `json:"attrs,omitempty"`
	// EncryptAttrs denotes whether to encrypt attribute values or not.
	// When set to true, each issued TCert in the batch will contain encrypted attribute values.
	EncryptAttrs bool `json:"encrypt_attrs,omitempty"`
	// Certificate Validity Period.  If specified, the value used
	// is the minimum of this value and the configured validity period
	// of the TCert manager.
	ValidityPeriod time.Duration `json:"validity_period,omitempty"`
	// The pre-key to be used for key derivation.
	PreKey string `json:"prekey"`
}

// GetBatchResponse is the response from the GetBatch API
type GetBatchResponse struct {
	ID     *big.Int  `json:"id"`
	TS     time.Time `json:"ts"`
	Key    []byte    `json:"key"`
	TCerts []TCert   `json:"tcerts"`
}

// TCert encapsulates a signed transaction certificate and optionally a map of keys
type TCert struct {
	Cert []byte            `json:"cert"`
	Keys map[string][]byte `json:"keys,omitempty"` //base64 encoded string as value
}

// Attribute is a single attribute name and value
type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
