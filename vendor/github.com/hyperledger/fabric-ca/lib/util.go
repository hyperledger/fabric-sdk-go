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
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/util"
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
		return nil, fmt.Errorf("buffer was neither PEM nor DER encoding: %s", err)
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

// UnmarshalConfig will use the viperunmarshal workaround to unmarshal a
// configuration file into a struct
func UnmarshalConfig(config interface{}, vp *viper.Viper, configFile string, server, viperIssue327WorkAround bool) error {
	vp.SetConfigFile(configFile)
	err := vp.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Failed to read config file: %s", err)
	}

	// Unmarshal the config into 'caConfig'
	// When viper bug https://github.com/spf13/viper/issues/327 is fixed
	// and vendored, the work around code can be deleted.
	if viperIssue327WorkAround {
		sliceFields := []string{
			"csr.hosts",
			"tls.clientauth.certfiles",
			"ldap.tls.certfiles",
			"db.tls.certfiles",
			"cafiles",
			"intermediate.tls.certfiles",
		}
		err = util.ViperUnmarshal(config, sliceFields, vp)
		if err != nil {
			return fmt.Errorf("Incorrect format in file '%s': %s", configFile, err)
		}
		if server {
			serverCfg := config.(*ServerConfig)
			err = util.ViperUnmarshal(&serverCfg.CAcfg, sliceFields, vp)
			if err != nil {
				return fmt.Errorf("Incorrect format in file '%s': %s", configFile, err)
			}
		}
	} else {
		err = vp.Unmarshal(config)
		if err != nil {
			return fmt.Errorf("Incorrect format in file '%s': %s", configFile, err)
		}
		if server {
			serverCfg := config.(*ServerConfig)
			err = vp.Unmarshal(&serverCfg.CAcfg)
			if err != nil {
				return fmt.Errorf("Incorrect format in file '%s': %s", configFile, err)
			}
		}
	}
	return nil
}

// GetAttrValue searches 'attrs' for the attribute with name 'name' and returns
// its value, or "" if not found.
func GetAttrValue(attrs []api.Attribute, name string) string {
	for _, attr := range attrs {
		if attr.Name == name {
			return attr.Value
		}
	}
	return ""
}

func getMaxEnrollments(userMaxEnrollments int, caMaxEnrollments int) (int, error) {
	log.Debugf("Max enrollment value verification - User specified max enrollment: %d, CA max enrollment: %d", userMaxEnrollments, caMaxEnrollments)
	if userMaxEnrollments < -1 {
		return 0, fmt.Errorf("Max enrollment in registration request may not be less than -1, but was %d", userMaxEnrollments)
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
		return 0, errors.New("Registration is disabled")
	default:
		switch userMaxEnrollments {
		case -1:
			// User requested infinite enrollments is not allowed
			return 0, errors.New("Registration for infinite enrollments is not allowed")
		case 0:
			// User is requesting the current CA maximum
			return caMaxEnrollments, nil
		default:
			// User is requesting a specific positive value; make sure it doesn't exceed the CA maximum.
			if userMaxEnrollments > caMaxEnrollments {
				return 0, fmt.Errorf("Requested enrollments (%d) exceeds maximum allowable enrollments (%d)",
					userMaxEnrollments, caMaxEnrollments)
			}
			// otherwise, use the requested maximum
			return userMaxEnrollments, nil
		}
	}
}
