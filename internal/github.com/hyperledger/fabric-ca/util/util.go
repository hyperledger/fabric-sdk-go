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

package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/lib/caerrors"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ocsp"
)

var (
	rnd = mrand.NewSource(time.Now().UnixNano())
	// ErrNotImplemented used to return errors for functions not implemented
	ErrNotImplemented = errors.New("NOT YET IMPLEMENTED")
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RevocationReasonCodes is a map between string reason codes to integers as defined in RFC 5280
var RevocationReasonCodes = map[string]int{
	"unspecified":          ocsp.Unspecified,
	"keycompromise":        ocsp.KeyCompromise,
	"cacompromise":         ocsp.CACompromise,
	"affiliationchanged":   ocsp.AffiliationChanged,
	"superseded":           ocsp.Superseded,
	"cessationofoperation": ocsp.CessationOfOperation,
	"certificatehold":      ocsp.CertificateHold,
	"removefromcrl":        ocsp.RemoveFromCRL,
	"privilegewithdrawn":   ocsp.PrivilegeWithdrawn,
	"aacompromise":         ocsp.AACompromise,
}

// SecretTag to tag a field as secret as in password, token
const SecretTag = "mask"

// URLRegex is the regular expression to check if a value is an URL
var URLRegex = regexp.MustCompile("(ldap|http)s*://(\\S+):(\\S+)@")

//ECDSASignature forms the structure for R and S value for ECDSA
type ECDSASignature struct {
	R, S *big.Int
}

// RandomString returns a random string
func RandomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, rnd.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rnd.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// RemoveQuotes removes outer quotes from a string if necessary
func RemoveQuotes(str string) string {
	if str == "" {
		return str
	}
	if (strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'")) ||
		(strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"")) {
		str = str[1 : len(str)-1]
	}
	return str
}

// ReadFile reads a file
func ReadFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}

// WriteFile writes a file
func WriteFile(file string, buf []byte, perm os.FileMode) error {
	dir := path.Dir(file)
	// Create the directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Wrapf(err, "Failed to create directory '%s' for file '%s'", dir, file)
		}
	}
	return ioutil.WriteFile(file, buf, perm)
}

// FileExists checks to see if a file exists
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Marshal to bytes
func Marshal(from interface{}, what string) ([]byte, error) {
	buf, err := json.Marshal(from)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to marshal %s", what)
	}
	return buf, nil
}

// Unmarshal from bytes
func Unmarshal(from []byte, to interface{}, what string) error {
	err := json.Unmarshal(from, to)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal %s", what)
	}
	return nil
}

// CreateToken creates a JWT-like token.
// In a normal JWT token, the format of the token created is:
//      <algorithm,claims,signature>
// where each part is base64-encoded string separated by a period.
// In this JWT-like token, there are two differences:
// 1) the claims section is a certificate, so the format is:
//      <certificate,signature>
// 2) the signature uses the private key associated with the certificate,
//    and the signature is across both the certificate and the "body" argument,
//    which is the body of an HTTP request, though could be any arbitrary bytes.
// @param cert The pem-encoded certificate
// @param key The pem-encoded key
// @param method http method of the request
// @param uri URI of the request
// @param body The body of an HTTP request
func CreateToken(csp bccsp.BCCSP, cert []byte, key bccsp.Key, method, uri string, body []byte) (string, error) {
	x509Cert, err := GetX509CertificateFromPEM(cert)
	if err != nil {
		return "", err
	}
	publicKey := x509Cert.PublicKey

	var token string

	//The RSA Key Gen is commented right now as there is bccsp does
	switch publicKey.(type) {
	/*
		case *rsa.PublicKey:
			token, err = GenRSAToken(csp, cert, key, body)
			if err != nil {
				return "", err
			}
	*/
	case *ecdsa.PublicKey:
		token, err = GenECDSAToken(csp, cert, key, method, uri, body)
		if err != nil {
			return "", err
		}
	}
	return token, nil
}

//GenRSAToken signs the http body and cert with RSA using RSA private key
// @csp : BCCSP instance
/*
func GenRSAToken(csp bccsp.BCCSP, cert []byte, key []byte, body []byte) (string, error) {
	privKey, err := GetRSAPrivateKey(key)
	if err != nil {
		return "", err
	}
	b64body := B64Encode(body)
	b64cert := B64Encode(cert)
	bodyAndcert := b64body + "." + b64cert
	hash := sha512.New384()
	hash.Write([]byte(bodyAndcert))
	h := hash.Sum(nil)
	RSAsignature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA384, h[:])
	if err != nil {
		return "", errors.Wrap(err, "Failed to rsa.SignPKCS1v15")
	}
	b64sig := B64Encode(RSAsignature)
	token := b64cert + "." + b64sig

	return  token, nil
}
*/

//GenECDSAToken signs the http body and cert with ECDSA using EC private key
func GenECDSAToken(csp bccsp.BCCSP, cert []byte, key bccsp.Key, method, uri string, body []byte) (string, error) {
	b64body := B64Encode(body)
	b64cert := B64Encode(cert)
	b64uri := B64Encode([]byte(uri))
	payload := method + "." + b64uri + "." + b64body + "." + b64cert

	return genECDSAToken(csp, key, b64cert, payload)
}

func genECDSAToken(csp bccsp.BCCSP, key bccsp.Key, b64cert, payload string) (string, error) {
	digest, digestError := csp.Hash([]byte(payload), &bccsp.SHAOpts{})
	if digestError != nil {
		return "", errors.WithMessage(digestError, fmt.Sprintf("Hash failed on '%s'", payload))
	}

	ecSignature, err := csp.Sign(key, digest, nil)
	if err != nil {
		return "", errors.WithMessage(err, "BCCSP signature generation failure")
	}
	if len(ecSignature) == 0 {
		return "", errors.New("BCCSP signature creation failed. Signature must be different than nil")
	}

	b64sig := B64Encode(ecSignature)
	token := b64cert + "." + b64sig

	return token, nil

}

// VerifyToken verifies token signed by either ECDSA or RSA and
// returns the associated user ID
func VerifyToken(csp bccsp.BCCSP, token string, method, uri string, body []byte, compMode1_3 bool) (*x509.Certificate, error) {

	if csp == nil {
		return nil, errors.New("BCCSP instance is not present")
	}
	x509Cert, b64Cert, b64Sig, err := DecodeToken(token)
	if err != nil {
		return nil, err
	}
	sig, err := B64Decode(b64Sig)
	if err != nil {
		return nil, errors.WithMessage(err, "Invalid base64 encoded signature in token")
	}
	b64Body := B64Encode(body)
	b64uri := B64Encode([]byte(uri))
	sigString := method + "." + b64uri + "." + b64Body + "." + b64Cert

	pk2, err := csp.KeyImport(x509Cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, errors.WithMessage(err, "Public Key import into BCCSP failed with error")
	}
	if pk2 == nil {
		return nil, errors.New("Public Key Cannot be imported into BCCSP")
	}

	//bccsp.X509PublicKeyImportOpts
	//Using default hash algo
	digest, digestError := csp.Hash([]byte(sigString), &bccsp.SHAOpts{})
	if digestError != nil {
		return nil, errors.WithMessage(digestError, "Message digest failed")
	}

	valid, validErr := csp.Verify(pk2, sig, digest, nil)
	if compMode1_3 && !valid {
		log.Debugf("Failed to verify token based on new authentication header requirements: %s", err)
		sigString := b64Body + "." + b64Cert
		digest, digestError := csp.Hash([]byte(sigString), &bccsp.SHAOpts{})
		if digestError != nil {
			return nil, errors.WithMessage(digestError, "Message digest failed")
		}
		valid, validErr = csp.Verify(pk2, sig, digest, nil)
	}

	if validErr != nil {
		return nil, errors.WithMessage(validErr, "Token signature validation failure")
	}
	if !valid {
		return nil, errors.New("Token signature validation failed")
	}

	return x509Cert, nil
}

// DecodeToken extracts an X509 certificate and base64 encoded signature from a token
func DecodeToken(token string) (*x509.Certificate, string, string, error) {
	if token == "" {
		return nil, "", "", errors.New("Invalid token; it is empty")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, "", "", errors.New("Invalid token format; expecting 2 parts separated by '.'")
	}
	b64cert := parts[0]
	certDecoded, err := B64Decode(b64cert)
	if err != nil {
		return nil, "", "", errors.WithMessage(err, "Failed to decode base64 encoded x509 cert")
	}
	x509Cert, err := GetX509CertificateFromPEM(certDecoded)
	if err != nil {
		return nil, "", "", errors.WithMessage(err, "Error in parsing x509 certificate given block bytes")
	}
	return x509Cert, b64cert, parts[1], nil
}

//GetECPrivateKey get *ecdsa.PrivateKey from key pem
func GetECPrivateKey(raw []byte) (*ecdsa.PrivateKey, error) {
	decoded, _ := pem.Decode(raw)
	if decoded == nil {
		return nil, errors.New("Failed to decode the PEM-encoded ECDSA key")
	}
	ECprivKey, err := x509.ParseECPrivateKey(decoded.Bytes)
	if err == nil {
		return ECprivKey, nil
	}
	key, err2 := x509.ParsePKCS8PrivateKey(decoded.Bytes)
	if err2 == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return key.(*ecdsa.PrivateKey), nil
		case *rsa.PrivateKey:
			return nil, errors.New("Expecting EC private key but found RSA private key")
		default:
			return nil, errors.New("Invalid private key type in PKCS#8 wrapping")
		}
	}
	return nil, errors.Wrap(err2, "Failed parsing EC private key")
}

//GetRSAPrivateKey get *rsa.PrivateKey from key pem
func GetRSAPrivateKey(raw []byte) (*rsa.PrivateKey, error) {
	decoded, _ := pem.Decode(raw)
	if decoded == nil {
		return nil, errors.New("Failed to decode the PEM-encoded RSA key")
	}
	RSAprivKey, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
	if err == nil {
		return RSAprivKey, nil
	}
	key, err2 := x509.ParsePKCS8PrivateKey(decoded.Bytes)
	if err2 == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return nil, errors.New("Expecting RSA private key but found EC private key")
		case *rsa.PrivateKey:
			return key.(*rsa.PrivateKey), nil
		default:
			return nil, errors.New("Invalid private key type in PKCS#8 wrapping")
		}
	}
	return nil, errors.Wrap(err, "Failed parsing RSA private key")
}

// B64Encode base64 encodes bytes
func B64Encode(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

// B64Decode base64 decodes a string
func B64Decode(str string) (buf []byte, err error) {
	return base64.StdEncoding.DecodeString(str)
}

// StrContained returns true if 'str' is in 'strs'; otherwise return false
func StrContained(str string, strs []string) bool {
	for _, s := range strs {
		if strings.ToLower(s) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// IsSubsetOf returns an error if there is something in 'small' that
// is not in 'big'.  Both small and big are assumed to be comma-separated
// strings.  All string comparisons are case-insensitive.
// Examples:
// 1) IsSubsetOf('a,B', 'A,B,C') returns nil
// 2) IsSubsetOf('A,B,C', 'B,C') returns an error because A is not in the 2nd set.
func IsSubsetOf(small, big string) error {
	bigSet := strings.Split(big, ",")
	smallSet := strings.Split(small, ",")
	for _, s := range smallSet {
		if s != "" && !StrContained(s, bigSet) {
			return errors.Errorf("'%s' is not a member of '%s'", s, big)
		}
	}
	return nil
}

// HTTPRequestToString returns a string for an HTTP request for debuggging
func HTTPRequestToString(req *http.Request) string {
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	return fmt.Sprintf("%s %s\n%s",
		req.Method, req.URL, string(body))
}

// HTTPResponseToString returns a string for an HTTP response for debuggging
func HTTPResponseToString(resp *http.Response) string {
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	return fmt.Sprintf("statusCode=%d (%s)\n%s",
		resp.StatusCode, resp.Status, string(body))
}

// CreateClientHome will create a home directory if it does not exist
func CreateClientHome() (string, error) {
	log.Debug("CreateHome")
	home := filepath.Dir(GetDefaultConfigFile("fabric-ca-client"))

	if _, err := os.Stat(home); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(home, 0755)
			if err != nil {
				return "", err
			}
		}
	}
	return home, nil
}

// GetDefaultConfigFile gets the default path for the config file to display in usage message
func GetDefaultConfigFile(cmdName string) string {
	if cmdName == "fabric-ca-server" {
		var fname = fmt.Sprintf("%s-config.yaml", cmdName)
		// First check home env variables
		home := "."
		envs := []string{"FABRIC_CA_SERVER_HOME", "FABRIC_CA_HOME", "CA_CFG_PATH"}
		for _, env := range envs {
			envVal := os.Getenv(env)
			if envVal != "" {
				home = envVal
				break
			}
		}
		return path.Join(home, fname)
	}

	var fname = fmt.Sprintf("%s-config.yaml", cmdName)
	// First check home env variables
	var home string
	envs := []string{"FABRIC_CA_CLIENT_HOME", "FABRIC_CA_HOME", "CA_CFG_PATH"}
	for _, env := range envs {
		envVal := os.Getenv(env)
		if envVal != "" {
			home = envVal
			return path.Join(home, fname)
		}
	}

	return path.Join(os.Getenv("HOME"), ".fabric-ca-client", fname)
}

// GetX509CertificateFromPEMFile gets an X509 certificate from a file
func GetX509CertificateFromPEMFile(file string) (*x509.Certificate, error) {
	pemBytes, err := ReadFile(file)
	if err != nil {
		return nil, err
	}
	x509Cert, err := GetX509CertificateFromPEM(pemBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid certificate in '%s'", file)
	}
	return x509Cert, nil
}

// GetX509CertificateFromPEM get an X509 certificate from bytes in PEM format
func GetX509CertificateFromPEM(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, errors.New("Failed to PEM decode certificate")
	}
	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing certificate")
	}
	return x509Cert, nil
}

// GetX509CertificatesFromPEM returns X509 certificates from bytes in PEM format
func GetX509CertificatesFromPEM(pemBytes []byte) ([]*x509.Certificate, error) {
	chain := pemBytes
	var certs []*x509.Certificate
	for len(chain) > 0 {
		var block *pem.Block
		block, chain = pem.Decode(chain)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing certificate")
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

// GetCertificateDurationFromFile returns the validity duration for a certificate
// in a file.
func GetCertificateDurationFromFile(file string) (time.Duration, error) {
	cert, err := GetX509CertificateFromPEMFile(file)
	if err != nil {
		return 0, err
	}
	return GetCertificateDuration(cert), nil
}

// GetCertificateDuration returns the validity duration for a certificate
func GetCertificateDuration(cert *x509.Certificate) time.Duration {
	return cert.NotAfter.Sub(cert.NotBefore)
}

// GetEnrollmentIDFromPEM returns the EnrollmentID from a PEM buffer
func GetEnrollmentIDFromPEM(cert []byte) (string, error) {
	x509Cert, err := GetX509CertificateFromPEM(cert)
	if err != nil {
		return "", err
	}
	return GetEnrollmentIDFromX509Certificate(x509Cert), nil
}

// GetEnrollmentIDFromX509Certificate returns the EnrollmentID from the X509 certificate
func GetEnrollmentIDFromX509Certificate(cert *x509.Certificate) string {
	return cert.Subject.CommonName
}

// MakeFileAbs makes 'file' absolute relative to 'dir' if not already absolute
func MakeFileAbs(file, dir string) (string, error) {
	if file == "" {
		return "", nil
	}
	if filepath.IsAbs(file) {
		return file, nil
	}
	path, err := filepath.Abs(filepath.Join(dir, file))
	if err != nil {
		return "", errors.Wrapf(err, "Failed making '%s' absolute based on '%s'", file, dir)
	}
	return path, nil
}

// MakeFileNamesAbsolute makes all file names in the list absolute, relative to home
func MakeFileNamesAbsolute(files []*string, home string) error {
	for _, filePtr := range files {
		abs, err := MakeFileAbs(*filePtr, home)
		if err != nil {
			return err
		}
		*filePtr = abs
	}
	return nil
}

// Fatal logs a fatal message and exits
func Fatal(format string, v ...interface{}) {
	log.Fatalf(format, v...)
	os.Exit(1)
}

// GetUser returns username and password from CLI input
func GetUser(v *viper.Viper) (string, string, error) {
	var fabricCAServerURL string
	fabricCAServerURL = v.GetString("url")

	URL, err := url.Parse(fabricCAServerURL)
	if err != nil {
		return "", "", err
	}

	user := URL.User
	if user == nil {
		return "", "", errors.New("No username and password provided as part of the Fabric CA server URL")
	}

	eid := user.Username()
	if eid == "" {
		return "", "", errors.New("No username provided as part of URL")
	}

	pass, _ := user.Password()
	if pass == "" {
		return "", "", errors.New("No password provided as part of URL")
	}

	return eid, pass, nil
}

// GetSerialAsHex returns the serial number from certificate as hex format
func GetSerialAsHex(serial *big.Int) string {
	hex := fmt.Sprintf("%x", serial)
	return hex
}

// StructToString converts a struct to a string. If a field
// has a 'secret' tag, it is masked in the returned string
func StructToString(si interface{}) string {
	rval := reflect.ValueOf(si).Elem()
	tipe := rval.Type()
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	for i := 0; i < rval.NumField(); i++ {
		tf := tipe.Field(i)
		if !rval.FieldByName(tf.Name).CanSet() {
			continue // skip unexported fields
		}
		var fStr string
		tagv := tf.Tag.Get(SecretTag)
		if tagv == "password" || tagv == "username" {
			fStr = fmt.Sprintf("%s:**** ", tf.Name)
		} else if tagv == "url" {
			val, ok := rval.Field(i).Interface().(string)
			if ok {
				val = GetMaskedURL(val)
				fStr = fmt.Sprintf("%s:%v ", tf.Name, val)
			} else {
				fStr = fmt.Sprintf("%s:%v ", tf.Name, rval.Field(i).Interface())
			}
		} else {
			fStr = fmt.Sprintf("%s:%v ", tf.Name, rval.Field(i).Interface())
		}
		buffer.WriteString(fStr)
	}
	buffer.WriteString(" }")
	return buffer.String()
}

// GetMaskedURL returns masked URL. It masks username and password from the URL
// if present
func GetMaskedURL(url string) string {
	matches := URLRegex.FindStringSubmatch(url)

	// If there is a match, there should be four entries: 1 for
	// the match and 3 for submatches
	if len(matches) == 4 {
		matchIdxs := URLRegex.FindStringSubmatchIndex(url)
		matchStr := url[matchIdxs[0]:matchIdxs[1]]
		for idx := 2; idx < len(matches); idx++ {
			if matches[idx] != "" {
				matchStr = strings.Replace(matchStr, matches[idx], "****", 1)
			}
		}
		url = url[:matchIdxs[0]] + matchStr + url[matchIdxs[1]:len(url)]
	}
	return url
}

// NormalizeStringSlice checks for seperators
func NormalizeStringSlice(slice []string) []string {
	var normalizedSlice []string

	for _, item := range slice {
		// Remove surrounding brackets "[]" if specified
		if strings.HasPrefix(item, "[") && strings.HasSuffix(item, "]") {
			item = item[1 : len(item)-1]
		}
		// Split elements based on comma and add to normalized slice
		elems := strings.Split(item, ",")
		for _, elem := range elems {
			normalizedSlice = append(normalizedSlice, strings.TrimSpace(elem))
		}
	}
	return normalizedSlice
}

// NormalizeFileList provides absolute pathing for the list of files
func NormalizeFileList(files []string, homeDir string) ([]string, error) {
	var err error

	files = NormalizeStringSlice(files)

	for i, file := range files {
		files[i], err = MakeFileAbs(file, homeDir)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// CheckHostsInCert checks to see if host correctly inserted into certificate
func CheckHostsInCert(certFile string, hosts ...string) error {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to read certificate file at '%s'", certFile)
	}

	cert, err := GetX509CertificateFromPEM(certBytes)
	if err != nil {
		return errors.Wrap(err, "Failed to get certificate")
	}

	// combine DNSNames and IPAddresses from cert
	sans := cert.DNSNames
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	for _, host := range hosts {
		if !containsString(sans, host) {
			return errors.Errorf("Host '%s' was not found in the certificate in file '%s'", host, certFile)
		}
	}
	return nil
}

func containsString(list []string, item string) bool {
	for _, elem := range list {
		if elem == item {
			return true
		}
	}
	return false
}

// Read reads from Reader into a byte array
func Read(r io.Reader, data []byte) ([]byte, error) {
	j := 0
	for {
		n, err := r.Read(data[j:])
		j = j + n
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "Read failure")
		}

		if (n == 0 && j == len(data)) || j > len(data) {
			return nil, errors.New("Size of requested data is too large")
		}
	}

	return data[:j], nil
}

// Hostname name returns the hostname of the machine
func Hostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return hostname
}

// ValidateAndReturnAbsConf checks to see that there are no conflicts between the
// configuration file path and home directory. If no conflicts, returns back the absolute
// path for the configuration file and home directory.
func ValidateAndReturnAbsConf(configFilePath, homeDir, cmdName string) (string, string, error) {
	var err error
	var homeDirSet bool
	var configFileSet bool

	defaultConfig := GetDefaultConfigFile(cmdName) // Get the default configuration

	if configFilePath == "" {
		configFilePath = defaultConfig // If no config file path specified, use the default configuration file
	} else {
		configFileSet = true
	}

	if homeDir == "" {
		homeDir = filepath.Dir(defaultConfig) // If no home directory specified, use the default directory
	} else {
		homeDirSet = true
	}

	// Make the home directory absolute
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to get full path of config file")
	}
	homeDir = strings.TrimRight(homeDir, "/")

	if configFileSet && homeDirSet {
		log.Warning("Using both --config and --home CLI flags; --config will take precedence")
	}

	if configFileSet {
		configFilePath, err = filepath.Abs(configFilePath)
		if err != nil {
			return "", "", errors.Wrap(err, "Failed to get full path of configuration file")
		}
		return configFilePath, filepath.Dir(configFilePath), nil
	}

	configFile := filepath.Join(homeDir, filepath.Base(defaultConfig)) // Join specified home directory with default config file name
	return configFile, homeDir, nil
}

// GetSliceFromList will return a slice from a list
func GetSliceFromList(split string, delim string) []string {
	return strings.Split(strings.Replace(split, " ", "", -1), delim)
}

// ListContains looks through a comma separated list to see if a string exists
func ListContains(list, find string) bool {
	items := strings.Split(list, ",")
	for _, item := range items {
		item = strings.TrimPrefix(item, " ")
		if item == find {
			return true
		}
	}
	return false
}

//TODO:  move these out of production code

// FatalError will check to see if an error occured if so it will cause the test cases exit
func FatalError(t *testing.T, err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args)
	}
	if !assert.NoError(t, err, msg) {
		t.Fatal(msg)
	}
}

// ErrorContains will check to see if an error occurred, if so it will check that it contains
// the appropriate error message
func ErrorContains(t *testing.T, err error, contains, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args)
	}
	if assert.Error(t, err, msg) {
		assert.Contains(t, caerrors.Print(err), contains)
	}
}
