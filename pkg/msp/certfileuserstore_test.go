/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

var storePathRoot = "/tmp/testcertfileuserstore"
var storePath = filepath.Join(storePathRoot, "-certs")

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

func TestStore(t *testing.T) {

	cleanupTestPath(t, storePathRoot)
	defer cleanupTestPath(t, storePathRoot)

	store, err := NewCertFileUserStore(storePath)
	if err != nil {
		t.Fatalf("NewFileKeyValueStore failed [%s]", err)
	}
	cleanupTestPath(t, storePath)

	user1 := &msp.UserData{
		MSPID:                 "Org1",
		ID:                    "user1",
		EnrollmentCertificate: []byte(testCert1),
	}
	user2 := &msp.UserData{
		MSPID:                 "Org2",
		ID:                    "user2",
		EnrollmentCertificate: []byte(testCert2),
	}

	createStore(store, user1, t, user2)

	// Check key1, value1
	if err = checkStoreValue(store, user1, user1.EnrollmentCertificate); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user1.ID, err)
	}
	if err = store.Delete(msp.IdentityIdentifier{MSPID: user1.MSPID, ID: user1.ID}); err != nil {
		t.Fatalf("Delete %s failed [%s]", user1.ID, err)
	}
	if err = checkStoreValue(store, user2, user2.EnrollmentCertificate); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user2.ID, err)
	}
	if err = checkStoreValue(store, user1, nil); err != msp.ErrUserNotFound {
		t.Fatalf("checkStoreValue %s failed, expected core.ErrUserNotFound, got: %s", user1.ID, err)
	}

	// Check ke2, value2
	if err = checkStoreValue(store, user2, user2.EnrollmentCertificate); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", user2.ID, err)
	}
	if err = store.Delete(msp.IdentityIdentifier{MSPID: user2.MSPID, ID: user2.ID}); err != nil {
		t.Fatalf("Delete %s failed [%s]", user2.ID, err)
	}
	if err = checkStoreValue(store, user2, nil); err != msp.ErrUserNotFound {
		t.Fatalf("checkStoreValue %s failed, expected core.ErrUserNotFound, got: %s", user2.ID, err)
	}

	// Check non-existing key
	checkNonExistingKey(store, t)
}

func createStore(store *CertFileUserStore, user1 *msp.UserData, t *testing.T, user2 *msp.UserData) {
	if err := store.Store(user1); err != nil {
		t.Fatalf("Store %s failed [%s]", user1.ID, err)
	}
	if err := store.Store(user2); err != nil {
		t.Fatalf("Store %s failed [%s]", user2.ID, err)
	}
}

func checkNonExistingKey(store *CertFileUserStore, t *testing.T) {
	nonExistingKey := msp.IdentityIdentifier{
		MSPID: "Orgx",
		ID:    "userx",
	}
	_, err := store.Load(nonExistingKey)
	if err == nil || err != msp.ErrUserNotFound {
		t.Fatal("fetching value for non-existing key should return ErrUserNotFound")
	}
}

func TestCreateNewStore(t *testing.T) {

	_, err := NewCertFileUserStore("")
	if err == nil {
		t.Fatal("should return error for empty path")
	}
}

func checkStoreValue(store *CertFileUserStore, user *msp.UserData, expected []byte) error {
	userIdentifier := userIdentifier(user)
	storeKey := storeKeyFromUserIdentifier(userIdentifier)
	v, err := store.Load(userIdentifier)
	if err != nil {
		return err
	}
	if err = compare(v.EnrollmentCertificate, expected); err != nil {
		return err
	}
	file := filepath.Join(storePath, storeKey)
	if err != nil {
		return err
	}
	if expected == nil {
		_, err = os.Stat(file)
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
	if !bytes.Equal(vbytes, expected) {
		return errors.New("value from store comparison failed")
	}
	return nil
}

func userIdentifier(userData *msp.UserData) msp.IdentityIdentifier {
	return msp.IdentityIdentifier{MSPID: userData.MSPID, ID: userData.ID}
}
