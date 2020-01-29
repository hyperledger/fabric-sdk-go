/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"reflect"
	"testing"

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
		WithConfig(config.FromFile("test/connection-tls.json")),
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
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.(*gateway).options

	if options.CommitHandler != DefaultCommitHandlers.OrgAll {
		t.Fatal("DefaultCommitHandler not correctly initialized")
	}

	if options.Discovery != true {
		t.Fatal("Discovery not correctly initialized")
	}
}

func TestConnectWithSDK(t *testing.T) {
	sdk, err := fabsdk.New(config.FromFile("test/connection-tls.json"))

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

	options := gw.(*gateway).options

	if options.CommitHandler != DefaultCommitHandlers.OrgAll {
		t.Fatal("DefaultCommitHandler not correctly initialized")
	}

	if options.Discovery != true {
		t.Fatal("Discovery not correctly initialized")
	}
}

func TestConnectWithIdentity(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity(testCert, testPrivKey))

	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithIdentity(wallet, "user"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	if gw.(*gateway).options.Identity == nil {
		t.Fatal("Identity not set")
	}

	mspid := gw.(*gateway).options.Identity.Identifier().MSPID

	if !reflect.DeepEqual(mspid, "Org1MSP") {
		t.Fatalf("Incorrect mspid: %s", mspid)
	}
}

func TestConnectWithCommitHandler(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
		WithCommitHandler(DefaultCommitHandlers.OrgAny),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.(*gateway).options

	if options.CommitHandler != DefaultCommitHandlers.OrgAny {
		t.Fatal("CommitHandler not set correctly")
	}
}

func TestConnectWithDiscovery(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
		WithDiscovery(false),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.(*gateway).options

	if options.Discovery != false {
		t.Fatal("Discovery not set correctly")
	}
}

func TestConnectWithMultipleOptions(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
		WithCommitHandler(DefaultCommitHandlers.OrgAny),
		WithDiscovery(false),
	)
	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	options := gw.(*gateway).options

	if options.Discovery != false {
		t.Fatal("Discovery not set correctly")
	}

	if options.CommitHandler != DefaultCommitHandlers.OrgAny {
		t.Fatal("CommitHandler not set correctly")
	}
}

func TestGetSdk(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	if gw.getSdk() != gw.(*gateway).sdk {
		t.Fatal("getSdk() not returning the correct object")
	}
}

func TestGetOrg(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
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

func TestGetPeersForOrg(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	peers, err := gw.getPeersForOrg("Org1")

	if err != nil {
		t.Fatalf("Failed to get peers for org: %s", err)
	}

	if reflect.DeepEqual(peers, [1]string{"peer0.org1.example.com"}) {
		t.Fatalf("GetPeersForOrg(Org1) returns: %s", peers)
	}

	peers, err = gw.getPeersForOrg("Org2")

	if reflect.DeepEqual(peers, [1]string{"peer0.org2.example.com"}) {
		t.Fatalf("GetPeersForOrg(Org1) returns: %s", peers)
	}

	peers, err = gw.getPeersForOrg("Org3")

	if err == nil {
		t.Fatal("GetPeersForOrg(Org3) should have returned error")
	}
}

func TestGetNetwork(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("user", NewX509Identity(testCert, testPrivKey))

	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithIdentity(wallet, "user"))

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	_, err = gw.GetNetwork("mychannel")
	if err == nil {
		t.Fatalf("Failed to get network: %s", err)
	}

	// name := nw.GetName()
	// if name != "mychannel" {
	// 	t.Fatalf("GetNetwork() not returning the correct named network: %s", name)
	// }
}

func TestClose(t *testing.T) {
	gw, err := Connect(
		WithConfig(config.FromFile("test/connection-tls.json")),
		WithUser("user1"),
	)

	if err != nil {
		t.Fatalf("Failed to create gateway: %s", err)
	}

	gw.Close()
}
