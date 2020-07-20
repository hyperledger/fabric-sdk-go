/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/keyutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

var pathRevokeCaRoot = filepath.Join(metadata.GetProjectPath(), metadata.CryptoConfigPath, "peerOrganizations/org1.example.com/ca/")
var pathParentCert = filepath.Join(metadata.GetProjectPath(), metadata.CryptoConfigPath, "peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem")
var peerCertToBeRevoked = filepath.Join(metadata.GetProjectPath(), metadata.CryptoConfigPath, "peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/signcerts/peer0.org1.example.com-cert.pem")
var newCRL string
var revokedCert string

//use this one to sign CRL
var orgTwoCA string

func TestMain(m *testing.M) {
	crl, e := generateCRL(peerCertToBeRevoked, pathRevokeCaRoot, pathParentCert)
	if e != nil {
		panic(fmt.Sprintf("error generating CRL for test : %s", e))
	}
	newCRL = string(crl)

	raw, err := ioutil.ReadFile(peerCertToBeRevoked)
	if err != nil {
		panic(fmt.Sprintf("failed to read cert to be revoked : %s", e))
	}
	revokedCert = string(raw)

	raw, err = ioutil.ReadFile(pathParentCert)
	if err != nil {
		panic(fmt.Sprintf("failed to read cert to be revoked : %s", e))
	}
	orgTwoCA = string(raw)

	fmt.Println(newCRL, revokedCert)
	os.Exit(m.Run())
}

//TestCertSignedWithUnknownAuthority
func TestCertSignedWithUnknownAuthority(t *testing.T) {
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	fabCertPool, err := tls.NewCertPool(false)
	assert.Nil(t, err)
	endpointConfig := &mocks.MockConfig{CustomTLSCACertPool: fabCertPool}

	m, err := New(Context{Providers: ctx, EndpointConfig: endpointConfig}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	invalidSignatureCrt := []byte(invalidSignaturePem)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: invalidSignatureCrt}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	err = m.Validate(goodEndorser)
	if !strings.Contains(err.Error(), "certificate signed by unknown authority") {
		t.Fatal("Expected error:'supplied identity is not valid: x509: certificate signed by unknown authority'")
	}

}

//TestRevokedCertificate
func TestRevokedCertificate(t *testing.T) {

	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")

	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(orgTwoCA))}
	m, err := New(Context{Providers: ctx, EndpointConfig: mocks.NewMockEndpointConfig()}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(revokedCert)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	//Validation should return en error since created CRL contains
	//revoked certificate
	err = m.Validate(goodEndorser)
	assert.NotNil(t, err)
	if !strings.Contains(err.Error(), "The certificate has been revoked") {
		t.Fatalf("Expected error for revoked certificate, but got :%s", err)
	}

}

//TestExpiredCertificate
func TestCertificateDates(t *testing.T) {
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	if err != nil {
		t.Fatalf("Error %s", err)
	}
	fabCertPool, err := tls.NewCertPool(false)
	assert.Nil(t, err)
	endpointConfig := &mocks.MockConfig{CustomTLSCACertPool: fabCertPool}

	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(orgTwoCA))}
	m, err := New(Context{Providers: ctx, EndpointConfig: endpointConfig}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// Certificate is in the future
	cert := generateSelfSignedCert(t, time.Now().Add(24*time.Hour))
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(cert)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	err = m.Validate(goodEndorser)
	if !strings.Contains(err.Error(), "Certificate provided is not valid until later date") {
		t.Fatal("Expected error 'Certificate provided is not valid until later date'")
	}

	// Certificate is in the past
	cert = generateSelfSignedCert(t, time.Now().Add(-24*time.Hour))
	sID = &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(cert)}
	goodEndorser, err = proto.Marshal(sID)
	assert.Nil(t, err)
	err = m.Validate(goodEndorser)
	if !strings.Contains(err.Error(), "Certificate provided has expired") {
		t.Fatal("Expected error 'Certificate provided has expired'")
	}
}

func TestNewMembership(t *testing.T) {
	goodMSPID := "GoodMSP"
	badMSPID := "BadMSP"

	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")

	fabCertPool, err := tls.NewCertPool(false)
	assert.Nil(t, err)
	endpointConfig := &mocks.MockConfig{CustomTLSCACertPool: fabCertPool}

	// Test bad config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte("invalid"))}
	m, err := New(Context{Providers: ctx, EndpointConfig: endpointConfig}, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, m)

	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err = New(Context{Providers: ctx, EndpointConfig: endpointConfig}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(certPem)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)

	sID = &mb.SerializedIdentity{Mspid: badMSPID, IdBytes: []byte(certPem)}
	badEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)

	assert.Nil(t, m.Validate(goodEndorser))
	assert.NotNil(t, m.Validate(badEndorser))

	assert.Nil(t, m.Verify(goodEndorser, []byte("test"), []byte("test1")))
	assert.NotNil(t, m.Verify(badEndorser, []byte("test"), []byte("test1")))
}

func buildMSPConfig(name string, root []byte) *mb.MSPConfig {
	return &mb.MSPConfig{
		Type:   0,
		Config: marshalOrPanic(buildfabricMSPConfig(name, root)),
	}
}

func buildfabricMSPConfig(name string, root []byte) *mb.FabricMSPConfig {
	config := &mb.FabricMSPConfig{
		Name:                          name,
		Admins:                        [][]byte{},
		IntermediateCerts:             [][]byte{},
		OrganizationalUnitIdentifiers: []*mb.FabricOUIdentifier{},
		RootCerts:                     [][]byte{root},
		RevocationList:                [][]byte{[]byte(newCRL)},
		SigningIdentity:               nil,
		FabricNodeOus: &mb.FabricNodeOUs{
			Enable:              true,
			AdminOuIdentifier:   &mb.FabricOUIdentifier{OrganizationalUnitIdentifier: "admin"},
			ClientOuIdentifier:  &mb.FabricOUIdentifier{OrganizationalUnitIdentifier: "client"},
			PeerOuIdentifier:    &mb.FabricOUIdentifier{OrganizationalUnitIdentifier: "peer"},
			OrdererOuIdentifier: &mb.FabricOUIdentifier{OrganizationalUnitIdentifier: "client"},
		},
	}

	return config

}

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICJzCCAc2gAwIBAgIUHS1hbKgmtURco9FMkOTAVynQKCgwCgYIKoZIzj0EAwIw
cDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMQ8wDQYDVQQH
EwZEdXJoYW0xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMjAwNjE3MTA0ODAwWhcNMzUwNjE0MTA0ODAw
WjBwMQswCQYDVQQGEwJVUzEXMBUGA1UECBMOTm9ydGggQ2Fyb2xpbmExDzANBgNV
BAcTBkR1cmhhbTEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMT
Y2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABMbm
0K7nntIdKITqDvm0iA2IdXE30gcPijD+b5mNUbmkSTfekU7Y2Dn6+1mG9VRp0a6U
iFeo2l9nG2VZpODzaMGjRTBDMA4GA1UdDwEB/wQEAwIBBjASBgNVHRMBAf8ECDAG
AQH/AgEBMB0GA1UdDgQWBBQ8GSzQHrtf9oKIO89wav9TRCxYbTAKBggqhkjOPQQD
AgNIADBFAiEAgWNqI8SKF1EkDhEkNpRiBC/JH+IWdpXM4XBvRKvx3T0CICb+AKil
nalNCQP6jt4Z9Dvj19Xn/19D75PMhMms7sB0
-----END CERTIFICATE-----`

var certPem = `-----BEGIN CERTIFICATE-----
MIICrzCCAlWgAwIBAgIURgJ5whz9lvp3Fkk+xapPLxxgsgswCgYIKoZIzj0EAwIw
cDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMQ8wDQYDVQQH
EwZEdXJoYW0xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMjAwNjE3MTA0ODAwWhcNMzAwNjE1MTA1MzAw
WjBdMQswCQYDVQQGEwJVUzEXMBUGA1UECBMOTm9ydGggQ2Fyb2xpbmExFDASBgNV
BAoTC0h5cGVybGVkZ2VyMQ8wDQYDVQQLEwZjbGllbnQxDjAMBgNVBAMTBXVzZXIx
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAErrVSd6NA3YdlhsfRRLhSaeYeKA/h
fNKK7oBJUtLxUPXbHqxZ4u2s9UeuFrC5WVzHVAWvAHFKHkEEmPkWB48v/6OB3zCB
3DAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUF4jYTs6j
Vod3ahqMumslAHl+zXIwHwYDVR0jBBgwFoAUPBks0B67X/aCiDvPcGr/U0QsWG0w
IgYDVR0RBBswGYIXQW5kcmV3cy1NQlAtOS5icm9hZGJhbmQwWAYIKgMEBQYHCAEE
THsiYXR0cnMiOnsiaGYuQWZmaWxpYXRpb24iOiIiLCJoZi5FbnJvbGxtZW50SUQi
OiJ1c2VyMSIsImhmLlR5cGUiOiJjbGllbnQifX0wCgYIKoZIzj0EAwIDSAAwRQIh
ANqxuLKIvWlpSr+jklS+jcfb68hXqnC2zshR7y0aAEQCAiAHOsxSK8s42/ynDWYM
Rxtk8baO32vSrjao5ESHgew8Nw==
-----END CERTIFICATE-----`

var invalidSignaturePem = `-----BEGIN CERTIFICATE-----
MIICCzCCAbKgAwIBAgIQaiOerd7fYdLv3WOe3G7maTAKBggqhkjOPQQDAjBXMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEOMAwGA1UEChMFcm9vdDIxCzAJBgNVBAMTAmNhMB4XDTE3MTIyMTE3
MTE1NFoXDTI3MTIxOTE3MTE1NFowVTELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNh
bGlmb3JuaWExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xGTAXBgNVBAMTEHNpZ25j
ZXJ0LXJldm9rZWQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATkCUK/7PBlDVY6
IyYVdLIJaHjz5Bx3mTMwySYwUsDYU0zD0btx0EBAKjTMDiLqkC5dllaxrU4gzHxr
5hy99+zjo2IwYDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
DAYDVR0TAQH/BAIwADArBgNVHSMEJDAigCBdC72qnK+2ajHaE61O7EwMxTJgqgm7
evx+2WCfZMfxOjAKBggqhkjOPQQDAgNHADBEAiAnGpZxlGGG4GIRc3bmrIqtG7sz
O/7VzRFysxkwySQCNwIgedom1wB4w/W/p05tdh6YXo8kLrEOWUb9KMchm3iaKT8=
-----END CERTIFICATE-----`

type validity struct {
	NotBefore, NotAfter time.Time
}

type publicKeyInfo struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

type tbsCertificate struct {
	Raw                asn1.RawContent
	Version            int `asn1:"optional,explicit,default:0,tag:0"`
	SerialNumber       *big.Int
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Issuer             asn1.RawValue
	Validity           validity
	Subject            asn1.RawValue
	PublicKey          publicKeyInfo
	UniqueID           asn1.BitString   `asn1:"optional,tag:1"`
	SubjectUniqueID    asn1.BitString   `asn1:"optional,tag:2"`
	Extensions         []pkix.Extension `asn1:"optional,explicit,tag:3"`
}

type certificate struct {
	Raw                asn1.RawContent
	TBSCertificate     tbsCertificate
	SignatureAlgorithm pkix.AlgorithmIdentifier
	SignatureValue     asn1.BitString
}

// encodeCertToMemory returns a PEM representation of a certificate
func encodeCertToMemory(c certificate) string {
	b, err := asn1.Marshal(c)
	if err != nil {
		return fmt.Sprintf("Failed marshaling cert: %s", err)
	}
	block := &pem.Block{
		Bytes: b,
		Type:  "CERTIFICATE",
	}
	b = pem.EncodeToMemory(block)
	return string(b)
}

func generateSelfSignedCert(t *testing.T, now time.Time) string {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	// Generate a self-signed certificate
	testExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	//extraExtensionData := []byte("extra extension")
	commonName := "securekey.com"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"SK"},
			Country:      []string{"CA"},
		},
		NotBefore:             now.Add(-1 * time.Hour),
		NotAfter:              now.Add(1 * time.Hour),
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		SubjectKeyId:          []byte{1, 2, 3, 4},
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           testExtKeyUsage,
		UnknownExtKeyUsage:    testUnknownExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	certRaw, err := x509.CreateCertificate(rand.Reader, &template, &template, &k.PublicKey, k)
	assert.NoError(t, err)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	var newCert certificate
	_, err = asn1.Unmarshal(certRaw, &newCert)
	if err != nil {
		log.Fatalf("Failed to unmarshal certificate: %s", err)
	}
	return encodeCertToMemory(newCert)

}

func generateCRL(cerPath, pathRevokeCaRoot, pathParentCert string) ([]byte, error) {

	var parentKey string
	err := filepath.Walk(pathRevokeCaRoot, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "_sk") {
			parentKey = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	key, err := loadPrivateKey(parentKey)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load private key")
	}

	cert, err := loadCert(pathParentCert)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	certToBeRevoked, err := loadCert(cerPath)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	crlBytes, err := revokeCert(certToBeRevoked, cert, key)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to revoke cert")
	}

	return crlBytes, nil
}

func loadPrivateKey(path string) (interface{}, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key, err := keyutil.PEMToPrivateKey(raw, []byte(""))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func loadCert(path string) (*x509.Certificate, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}

func revokeCert(certToBeRevoked *x509.Certificate, parentCert *x509.Certificate, parentKey interface{}) ([]byte, error) {

	//Create a revocation record for the user
	clientRevocation := pkix.RevokedCertificate{
		SerialNumber:   certToBeRevoked.SerialNumber,
		RevocationTime: time.Now().UTC(),
	}

	curRevokedCertificates := []pkix.RevokedCertificate{clientRevocation}
	//Generate new CRL that includes the user's revocation
	newCrlList, err := parentCert.CreateCRL(rand.Reader, parentKey, curRevokedCertificates, time.Now().UTC(), time.Now().UTC().AddDate(20, 0, 0))
	if err != nil {
		return nil, err
	}

	//CRL pem Block
	crlPemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: newCrlList,
	}
	var crlBuffer bytes.Buffer
	//Encode it to X509 CRL pem format print it out
	err = pem.Encode(&crlBuffer, crlPemBlock)
	if err != nil {
		return nil, err
	}

	return crlBuffer.Bytes(), nil
}
