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
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"time"

	"math/big"
	"strconv"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp"
	cspsigner "github.com/hyperledger/fabric/bccsp/signer"
)

var (
	// TCertEncTCertIndex is the ASN1 object identifier of the TCert index.
	TCertEncTCertIndex = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7}

	// TCertEncEnrollmentID is the ASN1 object identifier of the enrollment id.
	TCertEncEnrollmentID = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 8}

	// TCertAttributesHeaders is the ASN1 object identifier of attributes header.
	TCertAttributesHeaders = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 9}

	// Padding for encryption.
	Padding = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

	// tcertSubject is the subject name placed in all generated TCerts
	tcertSubject = pkix.Name{CommonName: "Fabric Transaction Certificate"}
)

// LoadMgr is the constructor for a TCert manager given key and certificate file names
// @parameter caKeyFile is the file name for the CA's key
// @parameter caCertFile is the file name for the CA's cert
func LoadMgr(caKeyFile, caCertFile string, myCSP bccsp.BCCSP) (*Mgr, error) {
	_, caKey, caCert, err := util.GetSignerFromCertFile(caCertFile, myCSP)
	if err != nil && caCert == nil {
		return nil, fmt.Errorf("Failed to load cert [%s]", err)
	}
	if err != nil {
		// Fallback: attempt to read out of keyFile and import
		log.Debugf("No key found in BCCSP keystore, attempting fallback")
		key, err := util.ImportBCCSPKeyFromPEM(caKeyFile, myCSP, true)
		if err != nil {
			return nil, err
		}
		signer, err := cspsigner.New(myCSP, key)
		if err != nil {
			return nil, fmt.Errorf("Failed initializing CryptoSigner [%s]", err)
		}
		caKey = signer
	}

	return NewMgr(caKey, caCert)
}

// NewMgr is the constructor for a TCert manager given a key and an x509 certificate
// @parameter caKey is used for signing a certificate request
// @parameter caCert is used for extracting CA data to associate with issued certificates
func NewMgr(caKey interface{}, caCert *x509.Certificate) (*Mgr, error) {
	mgr := new(Mgr)
	mgr.CAKey = caKey
	mgr.CACert = caCert
	mgr.ValidityPeriod = time.Hour * 24 * 365 // default to 1 year
	mgr.MaxAllowedBatchSize = 1000
	return mgr, nil
}

// Mgr is the manager for the TCert library
type Mgr struct {
	// CAKey is used for signing a certificate request
	CAKey interface{}
	// CACert is used for extracting CA data to associate with issued certificates
	CACert *x509.Certificate
	// ValidityPeriod is the duration that the issued certificate will be valid
	// unless the user requests a shorter validity period.
	// The default value is 1 year.
	ValidityPeriod time.Duration
	// MaxAllowedBatchSize is the maximum number of TCerts which can be requested at a time.
	// The default value is 1000.
	MaxAllowedBatchSize int
}

// GetBatch gets a batch of TCerts
// @parameter req Is the TCert batch request
// @parameter ecert Is the enrollment certificate of the caller
func (tm *Mgr) GetBatch(req *GetBatchRequest, ecert *x509.Certificate) (*GetBatchResponse, error) {

	log.Debugf("GetBatch req=%+v", req)

	// Set numTCertsInBatch to the number of TCerts to get.
	// If 0 are requested, retrieve the maximum allowable;
	// otherwise, retrieve the number requested it not too many.
	var numTCertsInBatch int
	if req.Count == 0 {
		numTCertsInBatch = int(tm.MaxAllowedBatchSize)
	} else if req.Count <= tm.MaxAllowedBatchSize {
		numTCertsInBatch = int(req.Count)
	} else {
		return nil, fmt.Errorf("You may not request %d TCerts; the maximum is %d",
			req.Count, tm.MaxAllowedBatchSize)
	}

	// Certs are valid for the min of requested and configured max
	vp := tm.ValidityPeriod
	if req.ValidityPeriod > 0 && req.ValidityPeriod < vp {
		vp = req.ValidityPeriod
	}

	// Create a template from which to create all other TCerts.
	// Since a TCert is anonymous and unlinkable, do not include
	template := &x509.Certificate{
		Subject: tcertSubject,
	}
	template.NotBefore = time.Now()
	template.NotAfter = template.NotBefore.Add(vp)
	template.IsCA = false
	template.KeyUsage = x509.KeyUsageDigitalSignature
	template.SubjectKeyId = []byte{1, 2, 3, 4}

	// Generate nonce for TCertIndex
	nonce := make([]byte, 16) // 8 bytes rand, 8 bytes timestamp
	rand.Reader.Read(nonce[:8])

	pub := ecert.PublicKey.(*ecdsa.PublicKey)

	mac := hmac.New(sha512.New384, []byte(createHMACKey()))
	raw, _ := x509.MarshalPKIXPublicKey(pub)
	mac.Write(raw)
	kdfKey := mac.Sum(nil)

	var set []TCert

	for i := 0; i < numTCertsInBatch; i++ {
		tcertid, uuidError := GenerateIntUUID()
		if uuidError != nil {
			return nil, fmt.Errorf("Failure generating UUID: %s", uuidError)
		}
		// Compute TCertIndex
		tidx := []byte(strconv.Itoa(2*i + 1))
		tidx = append(tidx[:], nonce[:]...)
		tidx = append(tidx[:], Padding...)

		mac := hmac.New(sha512.New384, kdfKey)
		mac.Write([]byte{1})
		extKey := mac.Sum(nil)[:32]

		mac = hmac.New(sha512.New384, kdfKey)
		mac.Write([]byte{2})
		mac = hmac.New(sha512.New384, mac.Sum(nil))
		mac.Write(tidx)

		one := new(big.Int).SetInt64(1)
		k := new(big.Int).SetBytes(mac.Sum(nil))
		k.Mod(k, new(big.Int).Sub(pub.Curve.Params().N, one))
		k.Add(k, one)

		tmpX, tmpY := pub.ScalarBaseMult(k.Bytes())
		txX, txY := pub.Curve.Add(pub.X, pub.Y, tmpX, tmpY)
		txPub := ecdsa.PublicKey{Curve: pub.Curve, X: txX, Y: txY}

		// Compute encrypted TCertIndex
		encryptedTidx, encryptErr := CBCPKCS7Encrypt(extKey, tidx)
		if encryptErr != nil {
			return nil, encryptErr
		}

		extensions, ks, extensionErr := generateExtensions(tcertid, encryptedTidx, ecert, req)

		if extensionErr != nil {
			return nil, extensionErr
		}

		template.PublicKey = txPub
		template.Extensions = extensions
		template.ExtraExtensions = extensions
		template.SerialNumber = tcertid

		raw, err := x509.CreateCertificate(rand.Reader, template, tm.CACert, &txPub, tm.CAKey)
		if err != nil {
			return nil, fmt.Errorf("Failed in TCert x509.CreateCertificate: %s", err)
		}

		pem := ConvertDERToPEM(raw, "CERTIFICATE")

		set = append(set, TCert{pem, ks})
	}

	tcertID, randNumErr := GenNumber(big.NewInt(20))
	if randNumErr != nil {
		return nil, randNumErr
	}

	tcertResponse := &GetBatchResponse{tcertID, time.Now(), kdfKey, set}

	return tcertResponse, nil

}

/**
*  Create HMAC Key
*  returns HMAC String
 */
func createHMACKey() string {
	key := make([]byte, 49)
	rand.Reader.Read(key)
	var cooked = base64.StdEncoding.EncodeToString(key)
	return cooked
}

// Generate encrypted extensions to be included into the TCert (TCertIndex, EnrollmentID and attributes).
func generateExtensions(tcertid *big.Int, tidx []byte, enrollmentCert *x509.Certificate, batchRequest *GetBatchRequest) ([]pkix.Extension, map[string][]byte, error) {
	// For each TCert we need to store and retrieve to the user the list of Ks used to encrypt the EnrollmentID and the attributes.
	ks := make(map[string][]byte)
	attrs := batchRequest.Attrs
	extensions := make([]pkix.Extension, len(attrs))

	preK1 := batchRequest.PreKey
	mac := hmac.New(sha512.New384, []byte(preK1))
	mac.Write(tcertid.Bytes())
	preK0 := mac.Sum(nil)

	// Compute encrypted EnrollmentID
	mac = hmac.New(sha512.New384, preK0)
	mac.Write([]byte("enrollmentID"))
	enrollmentIDKey := mac.Sum(nil)[:32]

	enrollmentID := []byte(GetEnrollmentIDFromCert(enrollmentCert))
	enrollmentID = append(enrollmentID, Padding...)

	encEnrollmentID, err := CBCPKCS7Encrypt(enrollmentIDKey, enrollmentID)
	if err != nil {
		return nil, nil, err
	}

	// save k used to encrypt EnrollmentID
	ks["enrollmentId"] = enrollmentIDKey

	attributeIdentifierIndex := 9
	count := 0
	attributesHeader := make(map[string]int)

	// Append attributes to the extensions slice
	for i := 0; i < len(attrs); i++ {
		count++
		name := attrs[i].Name
		value := []byte(attrs[i].Value)

		// Save the position of the attribute extension on the header.
		attributesHeader[name] = count

		// Encrypt attribute if enabled
		if batchRequest.EncryptAttrs {
			mac = hmac.New(sha512.New384, preK0)
			mac.Write([]byte(name))
			attributeKey := mac.Sum(nil)[:32]

			value = append(value, Padding...)
			value, err = CBCPKCS7Encrypt(attributeKey, value)
			if err != nil {
				return nil, nil, err
			}

			// Save the key used to encrypt the attribute
			ks[name] = attributeKey
		}

		// Generate an ObjectIdentifier for the extension holding the attribute
		TCertEncAttributes := asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, attributeIdentifierIndex + count}

		// Add the attribute extension to the extensions array
		extensions[count-1] = pkix.Extension{Id: TCertEncAttributes, Critical: false, Value: value}
	}

	// Append the TCertIndex to the extensions
	extensions = append(extensions, pkix.Extension{Id: TCertEncTCertIndex, Critical: true, Value: tidx})

	// Append the encrypted EnrollmentID to the extensions
	extensions = append(extensions, pkix.Extension{Id: TCertEncEnrollmentID, Critical: false, Value: encEnrollmentID})

	// Append the attributes header if there was attributes to include in the TCert
	if len(attrs) > 0 {
		extensions = append(extensions, pkix.Extension{Id: TCertAttributesHeaders, Critical: false, Value: buildAttributesHeader(attributesHeader)})
	}

	return extensions, ks, nil
}

func buildAttributesHeader(attributesHeader map[string]int) []byte {
	var headerString string
	for k, v := range attributesHeader {
		headerString = headerString + k + "->" + strconv.Itoa(v) + "#"
	}
	return []byte(headerString)
}
