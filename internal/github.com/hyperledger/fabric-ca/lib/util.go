/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/log"
	"github.com/grantae/certinfo"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/caerrors"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var clientAuthTypes = map[string]tls.ClientAuthType{
	"noclientcert":               tls.NoClientCert,
	"requestclientcert":          tls.RequestClientCert,
	"requireanyclientcert":       tls.RequireAnyClientCert,
	"verifyclientcertifgiven":    tls.VerifyClientCertIfGiven,
	"requireandverifyclientcert": tls.RequireAndVerifyClientCert,
}

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
		return nil, errors.Wrap(err, "Buffer was neither PEM nor DER encoding")
	}
	return cert, err
}

// LoadPEMCertPool loads a pool of PEM certificates from list of files
func LoadPEMCertPool(certFiles []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	if len(certFiles) > 0 {
		for _, cert := range certFiles {
			log.Debugf("Reading cert file: %s", cert)
			pemCerts, err := ioutil.ReadFile(cert)
			if err != nil {
				return nil, err
			}

			log.Debugf("Appending cert %s to pool", cert)
			if !certPool.AppendCertsFromPEM(pemCerts) {
				return nil, errors.New("Failed to load cert pool")
			}
		}
	}

	return certPool, nil
}

// UnmarshalConfig unmarshals a configuration file
func UnmarshalConfig(config interface{}, vp *viper.Viper, configFile string,
	server bool) error {

	vp.SetConfigFile(configFile)
	err := vp.ReadInConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to read config file '%s'", configFile)
	}

	err = vp.Unmarshal(config)
	if err != nil {
		return errors.Wrapf(err, "Incorrect format in file '%s'", configFile)
	}
	if server {
		serverCfg := config.(*ServerConfig)
		err = vp.Unmarshal(&serverCfg.CAcfg)
		if err != nil {
			return errors.Wrapf(err, "Incorrect format in file '%s'", configFile)
		}
	}
	return nil
}

func getMaxEnrollments(userMaxEnrollments int, caMaxEnrollments int) (int, error) {
	log.Debugf("Max enrollment value verification - User specified max enrollment: %d, CA max enrollment: %d", userMaxEnrollments, caMaxEnrollments)
	if userMaxEnrollments < -1 {
		return 0, caerrors.NewHTTPErr(400, caerrors.ErrInvalidMaxEnroll, "Max enrollment in registration request may not be less than -1, but was %d", userMaxEnrollments)
	}
	switch caMaxEnrollments {
	case -1:
		if userMaxEnrollments == 0 {
			// The user is requesting the matching limit of the CA, so gets infinite
			return caMaxEnrollments, nil
		}
		// There is no CA max enrollment limit, so simply use the user requested value
		return userMaxEnrollments, nil
	case 0:
		// The CA max enrollment is 0, so registration is disabled.
		return 0, caerrors.NewHTTPErr(400, caerrors.ErrInvalidMaxEnroll, "Registration is disabled")
	default:
		switch userMaxEnrollments {
		case -1:
			// User requested infinite enrollments is not allowed
			return 0, caerrors.NewHTTPErr(400, caerrors.ErrInvalidMaxEnroll, "Registration for infinite enrollments is not allowed")
		case 0:
			// User is requesting the current CA maximum
			return caMaxEnrollments, nil
		default:
			// User is requesting a specific positive value; make sure it doesn't exceed the CA maximum.
			if userMaxEnrollments > caMaxEnrollments {
				return 0, caerrors.NewHTTPErr(400, caerrors.ErrInvalidMaxEnroll, "Requested enrollments (%d) exceeds maximum allowable enrollments (%d)", userMaxEnrollments, caMaxEnrollments)
			}
			// otherwise, use the requested maximum
			return userMaxEnrollments, nil
		}
	}
}

func addQueryParm(req *http.Request, name, value string) {
	url := req.URL.Query()
	url.Add(name, value)
	req.URL.RawQuery = url.Encode()
}

// IdentityDecoder decodes streams of data coming from the server into an Identity object
func IdentityDecoder(decoder *json.Decoder) error {
	var id api.IdentityInfo
	err := decoder.Decode(&id)
	if err != nil {
		return err
	}
	fmt.Printf("Name: %s, Type: %s, Affiliation: %s, Max Enrollments: %d, Attributes: %+v\n", id.ID, id.Type, id.Affiliation, id.MaxEnrollments, id.Attributes)
	return nil
}

// AffiliationDecoder decodes streams of data coming from the server into an Affiliation object
func AffiliationDecoder(decoder *json.Decoder) error {
	var aff api.AffiliationInfo
	err := decoder.Decode(&aff)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", aff.Name)
	return nil
}

// CertificateDecoder is needed to keep track of state, to see how many certificates
// have been returned for each enrollment ID.
type CertificateDecoder struct {
	certIDCount map[string]int
	storePath   string
}

// NewCertificateDecoder returns decoder for certificates
func NewCertificateDecoder(storePath string) *CertificateDecoder {
	cd := &CertificateDecoder{}
	cd.certIDCount = make(map[string]int)
	cd.storePath = storePath
	return cd
}

// CertificateDecoder decodes streams of data coming from the server
func (cd *CertificateDecoder) CertificateDecoder(decoder *json.Decoder) error {
	var cert certPEM
	err := decoder.Decode(&cert)
	if err != nil {
		return err
	}
	block, rest := pem.Decode([]byte(cert.PEM))
	if block == nil || len(rest) > 0 {
		return errors.New("Certificate decoding error")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	enrollmentID := certificate.Subject.CommonName
	if cd.storePath != "" {
		err = cd.StoreCert(enrollmentID, cd.storePath, []byte(cert.PEM))
		if err != nil {
			return err
		}
	}

	result, err := certinfo.CertificateText(certificate)
	if err != nil {
		return err
	}
	fmt.Printf(result)
	return nil
}

// StoreCert stores the certificate on the file system
func (cd *CertificateDecoder) StoreCert(enrollmentID, storePath string, cert []byte) error {
	cd.certIDCount[enrollmentID] = cd.certIDCount[enrollmentID] + 1

	err := os.MkdirAll(storePath, os.ModePerm)
	if err != nil {
		return err
	}

	var filePath string
	singleCertName := fmt.Sprintf("%s.pem", enrollmentID)
	switch cd.certIDCount[enrollmentID] {
	case 1: // Only one certificate returned, don't need to append number to certificate file name
		filePath = filepath.Join(storePath, singleCertName)
	case 2: // Two certificates returned, rename the old certificate to have number at the end
		err := os.Rename(filepath.Join(storePath, singleCertName), filepath.Join(storePath, fmt.Sprintf("%s-1.pem", enrollmentID)))
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("Failed to rename certificate: %s", singleCertName))
		}
		filePath = filepath.Join(storePath, fmt.Sprintf("%s-%d.pem", enrollmentID, cd.certIDCount[enrollmentID]))
	default:
		filePath = filepath.Join(storePath, fmt.Sprintf("%s-%d.pem", enrollmentID, cd.certIDCount[enrollmentID]))
	}

	err = ioutil.WriteFile(filePath, cert, 0644)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("Failed to store certificate at: %s", storePath))
	}

	return nil
}
