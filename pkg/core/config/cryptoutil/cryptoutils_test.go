/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptoutil

import (
	"testing"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
)

func TestGetPrivateKeyFromCert(t *testing.T) {

	cs := cryptosuite.GetDefault()

	_, err := GetPrivateKeyFromCert([]byte(""), cs)
	if err == nil {
		t.Fatal("Should have failed for not a cert")
	}

	_, err = GetPrivateKeyFromCert([]byte(emptyCert), cs)
	if err == nil {
		t.Fatal("Should have failed for empty cert")
	}

	_, err = GetPrivateKeyFromCert([]byte(malformedCert), cs)
	if err == nil {
		t.Fatal("Should have failed for malformed cert")
	}

	_, err = GetPrivateKeyFromCert([]byte(dsa2048), cs)
	if err == nil {
		t.Fatal("Should have failed since supported key public key type is [ECDSA, RSA]")
	}

	_, err = GetPrivateKeyFromCert([]byte(ecdsaCert), cs)
	if err == nil {
		t.Fatal("Should have failed since key is not imported")
	}
}

func TestX509KeyPair(t *testing.T) {

	cs := cryptosuite.GetDefault()

	// Not a cert
	_, err := X509KeyPair([]byte(""), nil, cs)
	if err == nil {
		t.Fatal("Should have failed for not a cert")
	}

	// Malformed cert
	_, err = X509KeyPair([]byte(malformedCert), nil, cs)
	if err == nil {
		t.Fatal("Should have failed for malformed cert")
	}

	// DSA Cert
	_, err = X509KeyPair([]byte(dsa2048), nil, cs)
	if err == nil {
		t.Fatal("Should have failed for DSA algorithm")
	}

	// ECSDA Cert
	_, err = X509KeyPair([]byte(ecdsaCert), nil, cs)
	if err != nil {
		t.Fatalf("Failed to load key pair: %s", err)
	}

	// RSA Cert
	_, err = X509KeyPair([]byte(rsaCert), nil, cs)
	if err != nil {
		t.Fatalf("Failed to load key pair: %s", err)
	}

}

func TestPrivateKey(t *testing.T) {

	// Private key without crypto suite
	pk := &PrivateKey{nil, nil, nil}

	publicKey := pk.Public()
	if publicKey != nil {
		t.Fatal("Key should be nil")
	}

	_, err := pk.Sign(nil, []byte("Hello"), nil)
	if err == nil {
		t.Fatal("Should have failed since crypto suite is nil")
	}

	// Private key without private key
	pk = &PrivateKey{cryptosuite.GetDefault(), nil, nil}
	_, err = pk.Sign(nil, []byte("Hello"), nil)
	if err == nil {
		t.Fatal("Should have failed since private key is nil")
	}

	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(keyPem), cryptosuite.GetDefault(), true)
	if err != nil {
		t.Fatalf("Failed to import private key from pem: %s", err)
	}

	// Proper private key (crypto suite and private key provided)
	pk = &PrivateKey{cryptosuite.GetDefault(), privateKey, nil}
	signed, err := pk.Sign(nil, []byte("Hello"), nil)
	if err != nil {
		t.Fatalf("Error signing message: %s", err)
	}

	if len(signed) == 0 {
		t.Fatal("Message not signed")
	}

}

// RSA Cert
const rsaCert = `-----BEGIN CERTIFICATE-----
MIIFdDCCBFygAwIBAgIQJ2buVutJ846r13Ci/ITeIjANBgkqhkiG9w0BAQwFADBv
MQswCQYDVQQGEwJTRTEUMBIGA1UEChMLQWRkVHJ1c3QgQUIxJjAkBgNVBAsTHUFk
ZFRydXN0IEV4dGVybmFsIFRUUCBOZXR3b3JrMSIwIAYDVQQDExlBZGRUcnVzdCBF
eHRlcm5hbCBDQSBSb290MB4XDTAwMDUzMDEwNDgzOFoXDTIwMDUzMDEwNDgzOFow
gYUxCzAJBgNVBAYTAkdCMRswGQYDVQQIExJHcmVhdGVyIE1hbmNoZXN0ZXIxEDAO
BgNVBAcTB1NhbGZvcmQxGjAYBgNVBAoTEUNPTU9ETyBDQSBMaW1pdGVkMSswKQYD
VQQDEyJDT01PRE8gUlNBIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MIICIjANBgkq
hkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAkehUktIKVrGsDSTdxc9EZ3SZKzejfSNw
AHG8U9/E+ioSj0t/EFa9n3Byt2F/yUsPF6c947AEYe7/EZfH9IY+Cvo+XPmT5jR6
2RRr55yzhaCCenavcZDX7P0N+pxs+t+wgvQUfvm+xKYvT3+Zf7X8Z0NyvQwA1onr
ayzT7Y+YHBSrfuXjbvzYqOSSJNpDa2K4Vf3qwbxstovzDo2a5JtsaZn4eEgwRdWt
4Q08RWD8MpZRJ7xnw8outmvqRsfHIKCxH2XeSAi6pE6p8oNGN4Tr6MyBSENnTnIq
m1y9TBsoilwie7SrmNnu4FGDwwlGTm0+mfqVF9p8M1dBPI1R7Qu2XK8sYxrfV8g/
vOldxJuvRZnio1oktLqpVj3Pb6r/SVi+8Kj/9Lit6Tf7urj0Czr56ENCHonYhMsT
8dm74YlguIwoVqwUHZwK53Hrzw7dPamWoUi9PPevtQ0iTMARgexWO/bTouJbt7IE
IlKVgJNp6I5MZfGRAy1wdALqi2cVKWlSArvX31BqVUa/oKMoYX9w0MOiqiwhqkfO
KJwGRXa/ghgntNWutMtQ5mv0TIZxMOmm3xaG4Nj/QN370EKIf6MzOi5cHkERgWPO
GHFrK+ymircxXDpqR+DDeVnWIBqv8mqYqnK8V0rSS527EPywTEHl7R09XiidnMy/
s1Hap0flhFMCAwEAAaOB9DCB8TAfBgNVHSMEGDAWgBStvZh6NLQm9/rEJlTvA73g
JMtUGjAdBgNVHQ4EFgQUu69+Aj36pvE8hI6t7jiY7NkyMtQwDgYDVR0PAQH/BAQD
AgGGMA8GA1UdEwEB/wQFMAMBAf8wEQYDVR0gBAowCDAGBgRVHSAAMEQGA1UdHwQ9
MDswOaA3oDWGM2h0dHA6Ly9jcmwudXNlcnRydXN0LmNvbS9BZGRUcnVzdEV4dGVy
bmFsQ0FSb290LmNybDA1BggrBgEFBQcBAQQpMCcwJQYIKwYBBQUHMAGGGWh0dHA6
Ly9vY3NwLnVzZXJ0cnVzdC5jb20wDQYJKoZIhvcNAQEMBQADggEBAGS/g/FfmoXQ
zbihKVcN6Fr30ek+8nYEbvFScLsePP9NDXRqzIGCJdPDoCpdTPW6i6FtxFQJdcfj
Jw5dhHk3QBN39bSsHNA7qxcS1u80GH4r6XnTq1dFDK8o+tDb5VCViLvfhVdpfZLY
Uspzgb8c8+a4bmYRBbMelC1/kZWSWfFMzqORcUx8Rww7Cxn2obFshj5cqsQugsv5
B5a6SE2Q8pTIqXOi6wZ7I53eovNNVZ96YUWYGGjHXkBrI/V5eu+MtWuLt29G9Hvx
PUsE2JOAWVrgQSQdso8VYFhH2+9uRv0V9dlfmrPb2LjkQLPNlzmuhbsdjrzch5vR
pu/xO28QOG8=
-----END CERTIFICATE-----`

const ecdsaCert = `-----BEGIN CERTIFICATE-----
MIICNjCCAdygAwIBAgIRAILSPmMB3BzoLIQGsFxwZr8wCgYIKoZIzj0EAwIwbDEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRowGAYDVQQDExF0bHNjYS5l
eGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBaMGwxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRQwEgYDVQQKEwtleGFtcGxlLmNvbTEaMBgGA1UEAxMRdGxzY2EuZXhh
bXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQfgKb4db53odNzdMXn
P5FZTZTFztOO1yLvCHDofSNfTPq/guw+YYk7ZNmhlhj8JHFG6dTybc9Qb/HOh9hh
gYpXo18wXTAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCBxaEP3nVHQx4r7tC+WO//vrPRM1t86SKN0s6XB
8LWbHTAKBggqhkjOPQQDAgNIADBFAiEA96HXwCsuMr7tti8lpcv1oVnXg0FlTxR/
SQtE5YgdxkUCIHReNWh/pluHTxeGu2jNCH1eh6o2ajSGeeizoapvdJbN
-----END CERTIFICATE-----`

// A garbage cert, which can be decoded into ill-formed cert
const malformedCert = `-----BEGIN CERTIFICATE-----
MIICATCCAWoCCQDidF+uNJR6czANBgkqhkiG9w0BAQUFADBFMQswCQYDVQQGEwJB
cyBQdHkgTHRkMB4XDTEyMDUwMTIyNTUxN1oXDTEzMDUwMTIyNTUxN1owRTELMAkG
A1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0
nodhz31kLEJoeLSkRmrv8l7exkGtO0REtIbirj9BBy64ZXVBE7khKGO2cnM8U7yj
w7Ntfh+IvCjZVA3d2XqHS3Pjrt4HmU/cGCONE8+NEXoqdzLUDPOix1qDDRBvXs81
IFdpZGdpdHMgUHR5IEx0ZDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAtpjl
KAV2qh6CYHZbdqixhDerjvJcD4Nsd7kExEZfHuECAwEAATANBgkqhkiG9w0BAQUF
AAOBgQCyOqs7+qpMrYCgL6OamDeCVojLoEp036PsnaYWf2NPmsVXdpYW40Foyyjp
VTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0
iv5otkxO5rxtGPv7o2J1eMBpCuSkydvoz3Ey/QwGqbBwEXQ4xYCgra336gqW2KQt
+LnDCkE8f5oBhCIisExc2i8PDvsRsY70g/2gs983ImJjVR8sDw==
-----END CERTIFICATE-----`

const emptyCert = `-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`

const keyPem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

// DSA Cert (type not supported)
const dsa2048 = `-----BEGIN CERTIFICATE-----
MIIFdTCCBTWgAwIBAgIJAJfyK94Nz1yPMAkGByqGSM44BAMwcDELMAkGA1UEBhMC
VVMxCzAJBgNVBAgTAkNBMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMwEQYDVQQK
EwpDbG91ZEZsYXJlMRQwEgYDVQQLEwtFbmdpbmVlcmluZzERMA8GA1UEAxMIVEVT
VCBEU0EwHhcNMTQwNDEyMDUwMTUyWhcNMjQwNDA5MDUwMTUyWjBwMQswCQYDVQQG
EwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xEzARBgNV
BAoTCkNsb3VkRmxhcmUxFDASBgNVBAsTC0VuZ2luZWVyaW5nMREwDwYDVQQDEwhU
RVNUIERTQTCCAzowggItBgcqhkjOOAQBMIICIAKCAQEA27xa+d5kAGDnxWkmZON9
rNHw73/M4cwKpKGMpxGEdMt+u7wBNt6tCH0v6dHo6726L6YUopxSzKahtzngxmT8
G/P2dcbiVUm6r2N1T7zX5+9tnwWYPcpexdX/mXUnoB1yNHSckDiG0k5EGlQTTFXm
g22aChvINIFaoEdR5IW3fOdiIX0zNWUBQ6eezsFuoy1anIb9WjOcCtmdvjPFtWdm
ZwGVfUp/CmJ+720GijTmsRB3dCqpQoxsFC+BtbtOtgX7pKPPsmICaYTgDqaY6Oc2
HyWvS6xnl5uaHa33sFz9EisIy48nUbajWnLN8+bqSb+iIbR9xKxe1NRUO5rvJtXC
mQIVAK2dU+z5hzWPAnuHp19T9y8JKm8JAoIBABk907ebpqMBTGcJ6kQiJshgmao2
zN3uUWiA3GCrdnq8JxumqoRTbsLQsxh+nvw24U8bK94NhhoUmQHfhl1GWb4seSUy
goN7NUOC9wDH9QfrEi9S9eUS07gsLQ4QEYJPbxC1Wu8MIXJ2RpuaSFh+TClsasaG
K54JOwNp4Nvh3CXYfwYL1Jtt9vOctN2tF8Rr9zQrSgZDdsJvr/cIprxhY8JB4D54
Bq77D4zzULz792TKTHXyjhObL4XQcXz8tWloYF/wC8ME64CpVOx6GveN/cy6rINL
G4T9epmheVDVmM33Mg2KgY+L+V3ll3QxBX/uygjuzCmK489u+OrP4cnXxJYDggEF
AAKCAQBWXTTRLajF7bMCe36hp4dzxBQ7kHilviT0yguAzkBcZBAyZlwzBRqJIN7u
rsWwzjsBFcEjyNmRH9kQBm6Ggr/uqCj1VBW1b3lKkyN6xqPssrShwdTZ2O9QzPFk
NZT7OR3ZolO6ydvBYrBNcrrhYC/3topt/44C+eOWfcySp8zOpHbjDRxx2vyln3JN
HJHILq1/0bT++e9JtkQAtKwteO4HoUiumZrfRLkovghHpwEgqVQcL2t/oRPeUsMW
UzcHtgJXD/xWkQ0BN36wR7Px/3qWBA1xZIfxSzrXV3vAY6MHqkN7sW55f4J4sVTM
EQnZh2IxGdAxVn2cVjyL6z8ofYUWo4HVMIHSMB0GA1UdDgQWBBTdWYYdSWrZr5eD
pf3QoSWZz0AbCDCBogYDVR0jBIGaMIGXgBTdWYYdSWrZr5eDpf3QoSWZz0AbCKF0
pHIwcDELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRMwEQYDVQQKEwpDbG91ZEZsYXJlMRQwEgYDVQQLEwtFbmdpbmVlcmlu
ZzERMA8GA1UEAxMIVEVTVCBEU0GCCQCX8iveDc9cjzAMBgNVHRMEBTADAQH/MAkG
ByqGSM44BAMDLwAwLAIUP2uvD9JJpn1e7YZ/5QJIjlXhFl8CFGfNcNS49a0bN4Md
2HTcWtoMC+5k
-----END CERTIFICATE-----`
