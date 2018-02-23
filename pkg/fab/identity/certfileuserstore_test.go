/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
	"github.com/pkg/errors"
)

var storePathRoot = "/tmp/testcertfileuserstore"
var storePath = path.Join(storePathRoot, "-certs")

var testPrivKey1 = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgp4qKKB0WCEfx7XiB
5Ul+GpjM1P5rqc6RhjD5OkTgl5OhRANCAATyFT0voXX7cA4PPtNstWleaTpwjvbS
J3+tMGTG67f+TdCfDxWYMpQYxLlE8VkbEzKWDwCYvDZRMKCQfv2ErNvb
-----END PRIVATE KEY-----`

var testCert1 = `-----BEGIN CERTIFICATE-----
MIICGTCCAcCgAwIBAgIRALR/1GXtEud5GQL2CZykkOkwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBbMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEfMB0GA1UEAwwWVXNlcjFAb3JnMS5leGFtcGxlLmNvbTBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABPIVPS+hdftwDg8+02y1aV5pOnCO9tIn
f60wZMbrt/5N0J8PFZgylBjEuUTxWRsTMpYPAJi8NlEwoJB+/YSs29ujTTBLMA4G
A1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQkMCKAIIeR0TY+iVFf
mvoEKwaToscEu43ZXSj5fTVJornjxDUtMAoGCCqGSM49BAMCA0cAMEQCID+dZ7H5
AiaiI2BjxnL3/TetJ8iFJYZyWvK//an13WV/AiARBJd/pI5A7KZgQxJhXmmR8bie
XdsmTcdRvJ3TS/6HCA==
-----END CERTIFICATE-----`

var testPrivKey2 = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg5Ahcehypz6IpAYy6
DtIf5zZsRjP4PtsmDhLbBJsXmD6hRANCAAR+YRAn8dFpDQDyvDA7JKPl5PoZenj3
m1KOnMry/mOZcnXnTIh2ASV4ss8VluzBcyHGAv7BCmxXxDkjcV9eybv8
-----END PRIVATE KEY-----`

var testCert2 = `-----BEGIN CERTIFICATE-----
MIICGjCCAcCgAwIBAgIRAIQkbh9nsGnLmDalAVlj8sUwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBbMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEfMB0GA1UEAwwWQWRtaW5Ab3JnMS5leGFtcGxlLmNvbTBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABH5hECfx0WkNAPK8MDsko+Xk+hl6ePeb
Uo6cyvL+Y5lydedMiHYBJXiyzxWW7MFzIcYC/sEKbFfEOSNxX17Ju/yjTTBLMA4G
A1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQkMCKAIIeR0TY+iVFf
mvoEKwaToscEu43ZXSj5fTVJornjxDUtMAoGCCqGSM49BAMCA0gAMEUCIQDVf8cL
NrfToiPzJpEFPGF+/8CpzOkl91oz+XJsvdgf5wIgI/e8mpvpplUQbU52+LejA36D
CsbWERvZPjR/GFEDEvc=
-----END CERTIFICATE-----`

func crypto(t *testing.T) core.CryptoSuite {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return(path.Join(storePathRoot, "-keys"))
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	c, err := sw.GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}
	return c
}
func TestStore(t *testing.T) {

	cleanup(t, storePathRoot)
	defer cleanup(t, storePathRoot)

	crypto := crypto(t)

	_, err := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(testPrivKey1), crypto, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed [%s]", err)
	}

	_, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(testPrivKey2), crypto, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed [%s]", err)
	}

	store, err := NewCertFileUserStore(storePath, crypto)
	if err != nil {
		t.Fatalf("NewFileKeyValueStore failed [%s]", err)
	}
	cleanup(t, storePath)
	err = store.Store(nil)
	if err == nil {
		t.Fatal("Store(nil) should throw error")
	}

	user1 := &User{
		mspID: "Org1",
		name:  "user1",
		enrollmentCertificate: []byte(testCert1),
	}
	user2 := &User{
		mspID: "Org2",
		name:  "user2",
		enrollmentCertificate: []byte(testCert2),
	}
	if err := store.Store(user1); err != nil {
		t.Fatalf("Store %s failed [%s]", user1.Name(), err)
	}
	if err := store.Store(user2); err != nil {
		t.Fatalf("Store %s failed [%s]", user2.Name(), err)
	}

	// Check key1, value1
	if err := checkStoreValue(store, user1, user1.EnrollmentCertificate()); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user1.Name(), err)
	}
	if err := store.Delete(user1); err != nil {
		t.Fatalf("Delete %s failed [%s]", user1.Name(), err)
	}
	if err := checkStoreValue(store, user2, user2.EnrollmentCertificate()); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user2.Name(), err)
	}
	if err := checkStoreValue(store, user1, nil); err != api.ErrUserNotFound {
		t.Fatalf("checkStoreValue %s failed, expected api.ErrUserNotFound, got: %v", user1.Name(), err)
	}

	// Check ke2, value2
	if err := checkStoreValue(store, user2, user2.EnrollmentCertificate()); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user2.Name(), err)
	}
	if err := store.Delete(user2); err != nil {
		t.Fatalf("Delete %s failed [%s]", user2.Name(), err)
	}
	if err := checkStoreValue(store, user2, nil); err != api.ErrUserNotFound {
		t.Fatalf("checkStoreValue %s failed, expected api.ErrUserNotFound, got: %v", user2.Name(), err)
	}

	// Check non-existing key
	nonExistingKey := api.UserKey{
		MspID: "Orgx",
		Name:  "userx",
	}
	_, err = store.Load(nonExistingKey)
	if err == nil || err != api.ErrUserNotFound {
		t.Fatal("fetching value for non-existing key should return ErrUserNotFound")
	}
}

func TestCreateNewStore(t *testing.T) {

	crypto := crypto(t)

	_, err := NewCertFileUserStore("", crypto)
	if err == nil {
		t.Fatal("should return error for empty path")
	}

	_, err = NewCertFileUserStore("mypath", nil)
	if err == nil {
		t.Fatal("should return error for nil cryptosuite")
	}
}

func cleanup(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}

func checkStoreValue(store *CertFileUserStore, user api.User, expected []byte) error {
	userKey := userKeyFromUser(user)
	storeKey := storeKeyFromUserKey(userKeyFromUser(user))
	v, err := store.Load(userKey)
	if err != nil {
		return err
	}
	if err = compare(v.EnrollmentCertificate(), expected); err != nil {
		return err
	}
	file := path.Join(storePath, storeKey)
	if err != nil {
		return err
	}
	if expected == nil {
		_, err := os.Stat(file)
		if err == nil {
			return fmt.Errorf("path shouldn't exist [%s]", file)
		}
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "stat file failed [%s]", file)
		}
		// Doesn't exist, OK
		return nil
	}
	certBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return compare(certBytes, expected)
}

func compare(v interface{}, expected []byte) error {
	var vbytes []byte
	var ok bool
	if v == nil {
		vbytes = nil
	} else {
		vbytes, ok = v.([]byte)
		if !ok {
			return errors.New("value is not []byte")
		}
	}
	if bytes.Compare(vbytes, expected) != 0 {
		return errors.New("value from store comparison failed")
	}
	return nil
}
