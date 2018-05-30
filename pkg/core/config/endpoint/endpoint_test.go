/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTLSEnabled(t *testing.T) {
	b := IsTLSEnabled("https://some.url/")
	if !b {
		t.Fatalf("IsTLSEnabled reutrned false for https://")
	}
	b = IsTLSEnabled("http://some.url/")
	if b {
		t.Fatalf("IsTLSEnabled reutrned true for http://")
	}
	b = IsTLSEnabled("grpcs://some.url/")
	if !b {
		t.Fatalf("IsTLSEnabled reutrned false for grpcs://")
	}
	b = IsTLSEnabled("grpc://some.url/")
	if b {
		t.Fatalf("IsTLSEnabled reutrned true for grpc://")
	}
}

func TestToAddress(t *testing.T) {
	u := ToAddress("grpcs://some.url")
	if strings.HasPrefix(u, "grpcs://") {
		t.Fatalf("expected url to have protocol trimmed")
	}
	u = ToAddress("grpc://some.url")
	if strings.HasPrefix(u, "grpc://") {
		t.Fatalf("expected url to have protocol trimmed")
	}
	u = ToAddress("https://some.url")
	if !strings.HasPrefix(u, "https://") {
		t.Fatalf("expected url to have kept https:// protocol as prefix")
	}
	u = ToAddress("http://some.url")
	if !strings.HasPrefix(u, "http://") {
		t.Fatalf("expected url to have kept http:// protocol as prefix")
	}
}

func TestAttemptSecured(t *testing.T) {
	b := AttemptSecured("http://some.url", true)
	if b {
		t.Fatalf("trying to attempt non secured with http:// but got true")
	}
	b = AttemptSecured("http://some.url", false)
	if b {
		t.Fatalf("trying to attempt non secured with http:// but got true")
	}
	b = AttemptSecured("grpc://some.url", true)
	if b {
		t.Fatalf("trying to attempt non secured with grpc:// but got true")
	}
	b = AttemptSecured("grpc://some.url", false)
	if b {
		t.Fatalf("trying to attempt secured with grpc:// but got true")
	}
	b = AttemptSecured("grpcs://some.url", true)
	if !b {
		t.Fatalf("trying to attempt non secured with grpcs://, but got false")
	}
	b = AttemptSecured("grpcs://some.url", false)
	if !b {
		t.Fatalf("trying to attempt secured with grpcs://, but got false")
	}
	b = AttemptSecured("some.url", true)
	if b {
		t.Fatalf("trying to attempt non secured with no protocol in url, but got true")
	}
	b = AttemptSecured("some.url", false)
	if !b {
		t.Fatalf("trying to attempt secured with no protocol in url, but got false")
	}
}

func TestTLSConfig_Bytes(t *testing.T) {
	pPem := `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`
	tlsConfig := TLSConfig{
		Path: "",
		Pem:  pPem,
	}

	e := tlsConfig.LoadBytes()
	if e != nil {
		t.Fatalf("error loading bytes for sample cert %s", e)
	}
	b := tlsConfig.Bytes()
	if len(b) == 0 {
		t.Fatalf("cert's Bytes() call returned empty byte array")
	}
	if len(b) != len([]byte(pPem)) {
		t.Fatalf("cert's Bytes() call returned different byte array for correct pem")
	}

	// test with empty pem
	tlsConfig.Pem = ""
	tlsConfig.Path = "../testdata/config_test.yaml"
	e = tlsConfig.LoadBytes()
	if e != nil {
		t.Fatalf("error loading bytes for empty pem cert %s", e)
	}
	b = tlsConfig.Bytes()
	if len(b) == 0 {
		t.Fatalf("cert's Bytes() call returned empty byte array")
	}

	// test with wrong pem
	tlsConfig.Pem = "wrongpemvalue"
	e = tlsConfig.LoadBytes()
	if e != nil {
		t.Fatalf("error loading bytes for wrong pem cert %s", e)
	}
	b = tlsConfig.Bytes()
	if len(b) != len([]byte("wrongpemvalue")) {
		t.Fatalf("cert's Bytes() call returned different byte array for wrong pem")
	}
}

func TestTLSConfig_TLSCertPostive(t *testing.T) {
	tlsConfig := &TLSConfig{
		Path: "../../../../test/fixtures/config/mutual_tls/client_sdk_go.pem",
		Pem:  "",
	}

	e := tlsConfig.LoadBytes()
	if e != nil {
		t.Fatalf("error loading certificate for sample cert path %s", e)
	}

	c, e := tlsConfig.TLSCert()
	if e != nil {
		t.Fatalf("error loading certificate for sample cert path %s", e)
	}
	if c == nil {
		t.Fatalf("cert's TLSCert() call returned empty certificate")
	}

	// test with both correct pem and path set
	tlsConfig.Path = "../../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	tlsConfig.Pem = `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`
	c, e = tlsConfig.TLSCert()
	if e != nil {
		t.Fatalf("error loading certificate for sample cert path and pem %s", e)
	}
	if c == nil {
		t.Fatalf("cert's TLSCert() call returned empty certificate")
	}

}

func TestTLSConfig_TLSCertNegative(t *testing.T) {

	// test with wrong path
	tlsConfig := &TLSConfig{
		Path: "dummy/path",
		Pem:  "",
	}
	c, e := tlsConfig.TLSCert()
	if e == nil {
		t.Fatal("expected error loading certificate for wrong cert path")
	}
	if c != nil {
		t.Fatalf("cert's TLSCert() call returned non empty certificate for wrong cert path")
	}

	// test with empty path and empty pem
	tlsConfig.Path = ""
	c, e = tlsConfig.TLSCert()
	if e == nil {
		t.Fatal("expected error loading certificate for empty cert path and empty pem")
	}
	if c != nil {
		t.Fatalf("cert's TLSCert() call returned non empty certificate for wrong cert path and empty pem")
	}

	// test with wrong pem and empty path
	tlsConfig.Path = ""
	tlsConfig.Pem = "wrongcertpem"
	c, e = tlsConfig.TLSCert()
	if e == nil {
		t.Fatalf("error loading certificate for empty cert path and and wrong pem %s", e)
	}
	if c != nil {
		t.Fatalf("cert's TLSCert() call returned non empty certificate")
	}

}

func TestTLSConfigBytes(t *testing.T) {

	// test with wrong path
	tlsConfig := &TLSConfig{
		Path: "../testdata/config_test.yaml",
		Pem:  "",
	}

	err := tlsConfig.LoadBytes()
	bytes1 := tlsConfig.Bytes()
	assert.Nil(t, err, "tlsConfig.Bytes supposed to succeed")
	assert.NotEmpty(t, bytes1, "supposed to get valid bytes")

	tlsConfig.Path = "../testdata/config_test_pem.yaml"
	bytes2 := tlsConfig.Bytes()
	assert.Nil(t, err, "tlsConfig.Bytes supposed to succeed")
	assert.NotEmpty(t, bytes2, "supposed to get valid bytes")

	//even after changing path, it should return previous bytes
	assert.Equal(t, bytes1, bytes2, "any update to tlsconfig path after load bytes call should not take effect")

	//call preload now
	err = tlsConfig.LoadBytes()
	bytes2 = tlsConfig.Bytes()
	assert.Nil(t, err, "tlsConfig.Bytes supposed to succeed")
	assert.NotEmpty(t, bytes2, "supposed to get valid bytes")

	//even after changing path, it should return previous bytes
	assert.NotEqual(t, bytes1, bytes2, "tlsConfig.LoadBytes() should refresh bytes")

}
