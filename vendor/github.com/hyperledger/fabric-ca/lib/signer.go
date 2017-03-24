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

package lib

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/api"
)

func newSigner(key, cert []byte, id *Identity) *Signer {
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
	name   string
	key    []byte
	cert   []byte
	id     *Identity
	client *Client
}

// Key returns the key bytes of this signer
func (s *Signer) Key() []byte {
	return s.key
}

// Cert returns the cert bytes of this signer
func (s *Signer) Cert() []byte {
	return s.cert
}

// RevokeSelf revokes only the certificate associated with this signer
func (s *Signer) RevokeSelf() error {
	log.Debugf("RevokeSelf %s", s.name)
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
