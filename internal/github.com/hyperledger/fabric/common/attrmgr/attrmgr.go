/*
Copyright IBM Corp. 2017 All Rights Reserved.

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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

/*
 * The attrmgr package contains utilities for managing attributes.
 * Attributes are added to an X509 certificate as an extension.
 */

package attrmgr

import (
	"encoding/asn1"
)

var (
	// AttrOID is the ASN.1 object identifier for an attribute extension in an
	// X509 certificate
	AttrOID = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7, 8, 1}
	// AttrOIDString is the string version of AttrOID
	AttrOIDString = "1.2.3.4.5.6.7.8.1"
)

// Attribute is a name/value pair
type Attribute interface {
	// GetName returns the name of the attribute
	GetName() string
	// GetValue returns the value of the attribute
	GetValue() string
}

// AttributeRequest is a request for an attribute
type AttributeRequest interface {
	// GetName returns the name of an attribute
	GetName() string
	// IsRequired returns true if the attribute is required
	IsRequired() bool
}

// Mgr is the attribute manager and is the main object for this package
type Mgr struct{}

// Attributes contains attribute names and values
type Attributes struct {
	Attrs map[string]string `json:"attrs"`
}
