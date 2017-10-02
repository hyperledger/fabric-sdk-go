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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"crypto/x509"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	log "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib/logbridge"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/attrmgr"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
)

func newSigner(key bccsp.Key, cert []byte, id *Identity) *Signer {
	return &Signer{
		key:    key,
		cert:   cert,
		id:     id,
		client: id.client,
	}
}

// Signer represents a signer
// Each identity may have multiple signers, currently one ecert and multiple tcerts
type Signer struct {
	key    bccsp.Key
	cert   []byte
	id     *Identity
	client *Client
}

// Key returns the key bytes of this signer
func (s *Signer) Key() bccsp.Key {
	return s.key
}

// Cert returns the cert bytes of this signer
func (s *Signer) Cert() []byte {
	return s.cert
}

// GetX509Cert returns the X509 certificate for this signer
func (s *Signer) GetX509Cert() (*x509.Certificate, error) {
	cert, err := util.GetX509CertificateFromPEM(s.cert)
	if err != nil {
		return nil, fmt.Errorf("Failed getting X509 certificate for '%s': %s", s.id.name, err)
	}
	return cert, nil
}

// RevokeSelf revokes only the certificate associated with this signer
func (s *Signer) RevokeSelf() error {
	log.Debugf("RevokeSelf %s", s.id.name)
	serial, aki, err := GetCertID(s.cert)
	if err != nil {
		return err
	}
	req := &api.RevocationRequest{
		Serial: serial,
		AKI:    aki,
	}
	return s.id.Revoke(req)
}

// Attributes returns the attributes that are in the certificate
func (s *Signer) Attributes() (*attrmgr.Attributes, error) {
	cert, err := s.GetX509Cert()
	if err != nil {
		return nil, fmt.Errorf("Failed getting attributes for '%s': %s", s.id.name, err)
	}
	attrs, err := attrmgr.New().GetAttributesFromCert(cert)
	if err != nil {
		return nil, fmt.Errorf("Failed getting attributes for '%s': %s", s.id.name, err)
	}
	return attrs, nil
}
