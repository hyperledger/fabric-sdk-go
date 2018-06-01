/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"

	"fmt"

	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/stretchr/testify/assert"
)

//TestCertSignedWithUnknownAuthority
func TestCertSignedWithUnknownAuthority(t *testing.T) {
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err := New(Context{Providers: ctx}, cfg)
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
	var err error
	goodMSPID := "GoodMSP"
	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")
	if err != nil {
		t.Fatalf("Error %s", err)
	}
	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(orgTwoCA))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, m)

	// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
	sID := &mb.SerializedIdentity{Mspid: goodMSPID, IdBytes: []byte(org2RevokedCert)}
	goodEndorser, err := proto.Marshal(sID)
	assert.Nil(t, err)
	//Validation should return en error since created CRL contains
	//revoked certificate
	err = m.Validate(goodEndorser)
	assert.NotNil(t, err)
	if !strings.Contains(err.Error(), "The certificate has been revoked") {
		t.Fatal("Expected error for revoked certificate")
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
	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(orgTwoCA))}
	m, err := New(Context{Providers: ctx}, cfg)
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

	// Test bad config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte("invalid"))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, m)

	// Test good config input
	cfg.MockMSPs = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
	m, err = New(Context{Providers: ctx}, cfg)
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
	}

	return config

}

var newCRL = `-----BEGIN X509 CRL-----
MIIBVDCB/AIBATAKBggqhkjOPQQDAjBzMQswCQYDVQQGEwJVUzETMBEGA1UECBMK
Q2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3Jn
Mi5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eub3JnMi5leGFtcGxlLmNvbRcNMTgw
MzE0MTYyNDM5WhcNMTgwMzE1MTYyNDM5WjAnMCUCFHvhRd6BdVtYCseQWUqLc0E0
srURFw0xODAzMTQxNjI0MzFaoC8wLTArBgNVHSMEJDAigCCiWSBNvWrbFMBabgLe
lFZ7Kp99vp5qBjunZ9Qr8LVEwTAKBggqhkjOPQQDAgNHADBEAiAVKHw2GK1vh+K1
udBElnT7c1VYay8iIVQeBAvlzq+a5wIgdW1s9So8MDwt627LaJXyrbs4ZdMmkOAn
HuI5WPVWHHQ=
-----END X509 CRL-----`

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

var org2RevokedCert = `-----BEGIN CERTIFICATE-----
MIIC8DCCApegAwIBAgIUe+FF3oF1W1gKx5BZSotzQTSytREwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzIuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzIuZXhhbXBsZS5jb20wHhcNMTgwMzE0MTYxOTAwWhcNMTkwMzE0MTYy
NDAwWjCBgDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQw
EgYDVQQKEwtIeXBlcmxlZGdlcjEaMAsGA1UECxMEcGVlcjALBgNVBAsTBG9yZzEx
JjAkBgNVBAMTHXBlZXItcmV2b2tlZC5vcmcyLmV4YW1wbGUuY29tMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAExo+Z4mjffIxHcKxPSIKr8RAhBsv0lra6SidAIFsz
MOjT7V47w5rBWbqbWJnOteuqkgcjra+yzPZsDbTY2WqwOqOB+jCB9zAOBgNVHQ8B
Af8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUlI8a5sgOnq7/uO9kiE0A
T5DDv4swKwYDVR0jBCQwIoAgolkgTb1q2xTAWm4C3pRWeyqffb6eagY7p2fUK/C1
RMEwFwYDVR0RBBAwDoIMNzg3MmVmNmI2OGRiMHIGCCoDBAUGBwgBBGZ7ImF0dHJz
Ijp7ImhmLkFmZmlsaWF0aW9uIjoib3JnMSIsImhmLkVucm9sbG1lbnRJRCI6InBl
ZXItcmV2b2tlZC5vcmcyLmV4YW1wbGUuY29tIiwiaGYuVHlwZSI6InBlZXIifX0w
CgYIKoZIzj0EAwIDRwAwRAIgA8RuyDIiS+XV8XhODkTNdqvP3DaJ+JMt8ZX4o1E1
fzECIE4DUQO2Dhp1ufZJqiym1AN61+PSIPOPj9n26nMkWNJ8
-----END CERTIFICATE-----`

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICQzCCAemgAwIBAgIQYZpqGmcswky9Iy1SHBIm8zAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMS5leGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBa
MHMxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMRkwFwYDVQQKExBvcmcxLmV4YW1wbGUuY29tMRwwGgYDVQQD
ExNjYS5vcmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
3WtPeUzseT9Wp9VUtkx6mF84plyhgTlI2pbrHa4wYKFSoQGmrt83px6Q5Qu9EmhW
1y6Fr8DxkHvvg1NX0bCGyaNfMF0wDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYG
BFUdJQAwDwYDVR0TAQH/BAUwAwEB/zApBgNVHQ4EIgQgh5HRNj6JUV+a+gQrBpOi
xwS7jdldKPl9NUmiuePENS0wCgYIKoZIzj0EAwIDSAAwRQIhALUmxdk1FP8uL1so
nLdU8D8CS2PW5DLbaMjhR1KVK3b7AiAD5vkgX1PXPRsFFYlbkp/Y+nDdDy+mk3N7
K7xCT/QO7Q==
-----END CERTIFICATE-----
`

var certPem = `-----BEGIN CERTIFICATE-----
MIICGDCCAb+gAwIBAgIQXOaCoTss6vG3zb/vRGWXuDAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMS5leGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBa
MFsxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMR8wHQYDVQQDExZwZWVyMC5vcmcxLmV4YW1wbGUuY29tMFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWXupBEBzx/Mnjz1hzIUeOGiVR4CV/7aS
Qv0aokqJanTD+x8MaavBNYbPUwwzUNc7c1Ydd12gUNHPnyj/r1YyuaNNMEswDgYD
VR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYDVR0jBCQwIoAgh5HRNj6JUV+a
+gQrBpOixwS7jdldKPl9NUmiuePENS0wCgYIKoZIzj0EAwIDRwAwRAIgT2CAHCtr
Ro1YX8QuD6dSZUAOmptC+xU5xhp+2MeY2BkCIHmLOMBU5KIyJ5Rah4QeiswJ/pge
0eiDDUjXWGduFy4x
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

//use this one to sign CRL
var orgTwoCA = `-----BEGIN CERTIFICATE-----
MIICRDCCAeqgAwIBAgIRANqpQ8r//fDaj4j6kuGJv8gwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzIuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzIuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBzMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEcMBoGA1UE
AxMTY2Eub3JnMi5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BLdwS4lO/WKTrsQt+Q2bLMIbntuM7Teg6fEXvKrpIHFNzaCsTlemFUVxQUugQfUA
/GGIaomaE1STfvbCtElCsSOjXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNVHSUECDAG
BgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEIKJZIE29atsUwFpuAt6U
Vnsqn32+nmoGO6dn1CvwtUTBMAoGCCqGSM49BAMCA0gAMEUCIQCH8+Vw0L38dv/v
9gWvLhQv69q2bS0FBiAFwR4M17Z/2QIgH5W6rmsItiwa7nD0eZyiGmCzzQXW01b4
5fDo4hNhETQ=
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
		IsCA: true,
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
