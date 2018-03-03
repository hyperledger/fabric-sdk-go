/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package membership

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/stretchr/testify/assert"
)

func TestNewMembership(t *testing.T) {
	goodMSPID := "GoodMSP"
	badMSPID := "BadMSP"

	ctx := mocks.NewMockProviderContext()
	cfg := mocks.NewMockChannelCfg("")

	// Test bad config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte("invalid"))}
	m, err := New(Context{Providers: ctx}, cfg)
	assert.NotNil(t, err)
	assert.Nil(t, m)

	// Test good config input
	cfg.MockMsps = []*mb.MSPConfig{buildMSPConfig(goodMSPID, []byte(validRootCA))}
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
	return &mb.FabricMSPConfig{
		Name:                          name,
		Admins:                        [][]byte{},
		IntermediateCerts:             [][]byte{},
		OrganizationalUnitIdentifiers: []*mb.FabricOUIdentifier{},
		RevocationList:                [][]byte{},
		RootCerts:                     [][]byte{root},
		SigningIdentity:               nil,
	}
}

func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

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

var keyPem = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg9e5CQn0/jnFMQj9o
xs12HqzJpUa4j7Sj3spkbL+3dFGhRANCAARZe6kEQHPH8yePPWHMhR44aJVHgJX/
tpJC/RqiSolqdMP7Hwxpq8E1hs9TDDNQ1ztzVh13XaBQ0c+fKP+vVjK5
-----END PRIVATE KEY-----
`
