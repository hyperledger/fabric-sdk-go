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
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/hyperledger/fabric-ca/util"
)

// GetCertID returns both the serial number and AKI (Authority Key ID) for the certificate
func GetCertID(bytes []byte) (string, string, error) {
	cert, err := BytesToX509Cert(bytes)
	if err != nil {
		return "", "", err
	}
	serial := util.GetSerialAsHex(cert.SerialNumber)
	aki := hex.EncodeToString(cert.AuthorityKeyId)
	return serial, aki, nil
}

// BytesToX509Cert converts bytes (PEM or DER) to an X509 certificate
func BytesToX509Cert(bytes []byte) (*x509.Certificate, error) {
	dcert, _ := pem.Decode(bytes)
	if dcert != nil {
		bytes = dcert.Bytes
	}
	cert, err := x509.ParseCertificate(bytes)
	if err != nil {
		return nil, fmt.Errorf("buffer was neither PEM nor DER encoding: %s", err)
	}
	return cert, err
}
