/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const testPrivKey string = `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQggkuKP0YNrbuilpFf
0F/I+3At9LZh6EysU8lVBuy+cregCgYIKoZIzj0DAQehRANCAAQ3NMOS6YpCyFKJ
jgKYCP6eQYUG91jdhoQK+8Ufhy0/V/CVdJj/Exe89yzAqKfLzb9tc6MuWOYLwPRD
sF3d8qsw
-----END PRIVATE KEY-----`

const testCert string = `-----BEGIN CERTIFICATE-----
MIICjzCCAjWgAwIBAgIUXtE0iOex19qEbY12PpU3Sig3/LswCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMjAwMTA3MTEzNjAwWhcNMjEwMTA2MTE0
MTAwWjBCMTAwDQYDVQQLEwZjbGllbnQwCwYDVQQLEwRvcmcxMBIGA1UECxMLZGVw
YXJ0bWVudDExDjAMBgNVBAMTBXVzZXIxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD
QgAENzTDkumKQshSiY4CmAj+nkGFBvdY3YaECvvFH4ctP1fwlXSY/xMXvPcswKin
y82/bXOjLljmC8D0Q7Bd3fKrMKOB1zCB1DAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0T
AQH/BAIwADAdBgNVHQ4EFgQUfi/LNRJof+w9YtBydB7kpget9eowKwYDVR0jBCQw
IoAga001uwQc4mqKCzZzSlqHrmd3JGYF3lbyxsEzYHvzmSEwaAYIKgMEBQYHCAEE
XHsiYXR0cnMiOnsiaGYuQWZmaWxpYXRpb24iOiJvcmcxLmRlcGFydG1lbnQxIiwi
aGYuRW5yb2xsbWVudElEIjoidXNlcjEiLCJoZi5UeXBlIjoiY2xpZW50In19MAoG
CCqGSM49BAMCA0gAMEUCIQCXMS8+ahDQZ5wHnWUcps9GH2uWG+qPO3LxTitCH/rs
owIgRo0pFBhgLXaJ9ECYR+gSNBDpIc5I/Fr7QL7iIleSQlY=
-----END CERTIFICATE-----`

func TestConnectIdentityInCcp(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	if gw == nil {
		t.Fatal("Failed to create gateway")
	}
}

func TestConnectNoOptions(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.options

	if options.Timeout != defaultTimeout {
		t.Fatal("Timeout not correctly initialized")
	}
}

func TestConnectWithBlockNum(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
		WithBlockNum(2),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.options

	if !options.FromBlockSet || options.FromBlock != 2 {
		t.Fatal("BlockNum not correctly initialized")
	}
}

func TestConnectWithSDK(t *testing.T) {
	sdk, err := fabsdk.New(config.FromFile("testdata/connection-tls.json"))

	if err != nil {
		t.Fatalf("Failed to create SDK: %s", err)
	}

	gw, err := Connect(
		WithSDK(sdk),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.options

	if options.Timeout != defaultTimeout {
		t.Fatal("Timeout not correctly initialized")
	}
}

func TestConnectWithIdentity(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity("testMSP", testCert, testPrivKey))

	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithIdentity(wallet, "user"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	if gw.options.Identity == nil {
		t.Fatal("Identity not set")
	}

	mspid := gw.options.Identity.Identifier().MSPID

	if !reflect.DeepEqual(mspid, "testMSP") {
		t.Fatalf("Incorrect mspid: %s", mspid)
	}
}

func TestConnectWithBadConfig(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity("testMSP", testCert, testPrivKey))

	badConfig := func(gw *Gateway) error {
		return errors.New("Failed Config")
	}

	_, err := Connect(
		badConfig,
		WithIdentity(wallet, "user"),
	)

	if err == nil {
		t.Fatal("Expected to fail to create gateway")
	}
}

func TestConnectWithBadIdentity(t *testing.T) {
	badIdentity := func(gw *Gateway) error {
		return errors.New("Failed Identity")
	}

	_, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		badIdentity,
	)

	if err == nil {
		t.Fatal("Expected to fail to create gateway")
	}
}

func TestConnectWithBadOption(t *testing.T) {
	badOption := func(gw *Gateway) error {
		return errors.New("Failed Option")
	}

	_, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
		badOption,
	)
	if err == nil {
		t.Fatal("Expected to fail to create gateway")
	}
}

func TestConnectWithTimout(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
		WithTimeout(20*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.options

	if options.Timeout != 20*time.Second {
		t.Fatal("Timeout not set correctly")
	}
}

func TestGetOrg(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	org := gw.getOrg()

	if org != "Org1" {
		t.Fatalf("getOrg() returns: %s", org)
	}
}

func TestGetNetworkWithIdentity(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity("msp", testCert, testPrivKey))

	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithIdentity(wallet, "user"))

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	if gw.options.Identity.PublicVersion() != gw.options.Identity {
		t.Fatal("Incorrect identity")
	}

	if string(gw.options.Identity.EnrollmentCertificate()) != testCert {
		t.Fatal("Incorrect identity certificate")
	}

	if gw.options.Identity.PrivateKey().Symmetric() {
		t.Fatal("Incorrect identity private key")
	}

	_, err = gw.GetNetwork("mychannel")
	if err == nil {
		t.Fatalf("Failed to get network: %s", err)
	}
}

func TestGetNetworkWithUser(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"))

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	_, err = gw.GetNetwork("mychannel")
	if err == nil {
		t.Fatalf("Failed to get network: %s", err)
	}
}

func TestClose(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	gw.Close()
}

func TestAsLocalhost(t *testing.T) {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")

	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity("msp", testCert, testPrivKey))

	gw, err := Connect(
		WithConfig(config.FromFile("testdata/connection-discovery.json")),
		WithIdentity(wallet, "user"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	_, err = gw.GetNetwork("mychannel")
	if err == nil {
		t.Fatalf("Failed to get network: %s", err)
	}

}
