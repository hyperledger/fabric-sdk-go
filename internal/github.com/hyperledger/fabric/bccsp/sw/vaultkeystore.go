/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package sw

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

// NewVaultBasedKeyStore instantiated a Vault-based key store at a given position.
// The key store can be encrypted if a non-empty password is specified.
// It can be also be set as read only. In this case, any store operation
// will be forbidden
func NewVaultBasedKeyStore(pwd []byte, path string) (bccsp.KeyStore, error) {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return nil, errors.New("VAULT_ADDR env is empty")
	}
	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultAddr == "" {
		return nil, errors.New("VAULT_TOKEN env is empty")
	}
	vaultSpace := os.Getenv("VAULT_SPACE")
	if vaultSpace == "VAULT_SPACE" {
		vaultSpace = "fabric/data"
	}
	v := NewVault(vaultAddr, vaultToken, vaultSpace)
	ks := &vaultBasedKeyStore{vault: v}
	ks.path = filepath.Join(vaultSpace, path)
	return ks, nil
}

// vaultBasedKeyStore is a Vault-based KeyStore.
// Each key is stored in a separated Vault path that contains the key's SKI
// and flags to identity the key's type. All the keys are stored in
// a folder whose path is provided at initialization time.
type vaultBasedKeyStore struct {
	path string

	readOnly bool
	isOpen   bool

	pwd []byte

	// Sync
	m sync.Mutex

	vault *Vault
}

// ReadOnly returns true if this KeyStore is read only, false otherwise.
// If ReadOnly is true then StoreKey will fail.
func (ks *vaultBasedKeyStore) ReadOnly() bool {
	return ks.readOnly
}

// GetKey returns a key object whose SKI is the one passed.
func (ks *vaultBasedKeyStore) GetKey(ski []byte) (bccsp.Key, error) {
	// Validate arguments
	if len(ski) == 0 {
		return nil, errors.New("invalid SKI. Cannot be of zero length")
	}

	suffix := ks.getSuffix(hex.EncodeToString(ski))

	switch suffix {
	case "key":
		// Load the key
		key, err := ks.loadKey(hex.EncodeToString(ski))
		if err != nil {
			return nil, fmt.Errorf("failed loading key [%x] [%s]", ski, err)
		}

		return &aesPrivateKey{key, false}, nil
	case "sk":
		// Load the private key
		key, err := ks.loadPrivateKey(hex.EncodeToString(ski))
		if err != nil {
			return nil, fmt.Errorf("failed loading secret key [%x] [%s]", ski, err)
		}

		switch k := key.(type) {
		case *ecdsa.PrivateKey:
			return &ecdsaPrivateKey{k, true}, nil
		default:
			return nil, errors.New("secret key type not recognized")
		}
	case "pk":
		// Load the public key
		key, err := ks.loadPublicKey(hex.EncodeToString(ski))
		if err != nil {
			return nil, fmt.Errorf("failed loading public key [%x] [%s]", ski, err)
		}

		switch k := key.(type) {
		case *ecdsa.PublicKey:
			return &ecdsaPublicKey{k}, nil
		default:
			return nil, errors.New("public key type not recognized")
		}
	default:
		return ks.searchKeystoreForSKI(ski)
	}
}

// StoreKey stores the key k in this KeyStore.
// If this KeyStore is read only then the method will fail.
func (ks *vaultBasedKeyStore) StoreKey(k bccsp.Key) (err error) {
	if ks.readOnly {
		return errors.New("read only KeyStore")
	}

	if k == nil {
		return errors.New("invalid key. It must be different from nil")
	}
	switch kk := k.(type) {
	case *ecdsaPrivateKey:
		err = ks.storePrivateKey(hex.EncodeToString(k.SKI()), kk.privKey)
		if err != nil {
			return fmt.Errorf("failed storing ECDSA private key [%s]", err)
		}

	case *ecdsaPublicKey:
		err = ks.storePublicKey(hex.EncodeToString(k.SKI()), kk.pubKey)
		if err != nil {
			return fmt.Errorf("failed storing ECDSA public key [%s]", err)
		}

	case *aesPrivateKey:
		err = ks.storeKey(hex.EncodeToString(k.SKI()), kk.privKey)
		if err != nil {
			return fmt.Errorf("failed storing AES key [%s]", err)
		}

	default:
		return fmt.Errorf("key type not reconigned [%s]", k)
	}

	return
}

func (ks *vaultBasedKeyStore) searchKeystoreForSKI(ski []byte) (k bccsp.Key, err error) {
	k, err = ks.vault.RecursiveSearch(ks.path, ski, ks.pwd)
	if k == nil {
		return nil, fmt.Errorf("key with SKI %x not found in %s", ski, ks.path)
	}

	return k, nil
}

func (ks *vaultBasedKeyStore) getSuffix(alias string) string {
	list, err := ks.vault.Client.Logical().List(ks.path)
	if err != nil || list == nil {
		return ""
	}

	keys, ok := list.Data["keys"].([]interface{})
	if !ok {
		return ""
	}
	for _, k := range keys {
		key := k.(string)
		if strings.HasPrefix(key, alias) {
			if strings.HasSuffix(key, "sk") {
				return "sk"
			}
			if strings.HasSuffix(key, "pk") {
				return "pk"
			}
			if strings.HasSuffix(key, "key") {
				return "key"
			}
			break
		}
	}
	return ""
}

func (ks *vaultBasedKeyStore) storePrivateKey(alias string, privateKey interface{}) error {
	rawKey, err := privateKeyToPEM(privateKey, ks.pwd)
	if err != nil {
		logger.Errorf("Failed converting private key to PEM [%s]: [%s]", alias, err)
		return err
	}

	data := map[string]interface{}{
		"data": rawKey,
	}
	_, err = ks.vault.Client.Logical().Write(filepath.Join(ks.path, ks.getPathForAlias(alias, "sk")), data)
	if err != nil {
		logger.Errorf("Failed storing private key [%s]: [%s]", alias, err)
		return err
	}

	return nil
}

func (ks *vaultBasedKeyStore) storePublicKey(alias string, publicKey interface{}) error {
	rawKey, err := publicKeyToPEM(publicKey, ks.pwd)
	if err != nil {
		logger.Errorf("Failed converting public key to PEM [%s]: [%s]", alias, err)
		return err
	}

	data := map[string]interface{}{"data": rawKey}
	_, err = ks.vault.Client.Logical().Write(ks.getPathForAlias(alias, "pk"), data)
	if err != nil {
		logger.Errorf("Failed storing private key [%s]: [%s]", alias, err)
		return err
	}

	return nil
}

func (ks *vaultBasedKeyStore) storeKey(alias string, key []byte) error {
	pem, err := aesToEncryptedPEM(key, ks.pwd)
	if err != nil {
		logger.Errorf("Failed converting key to PEM [%s]: [%s]", alias, err)
		return err
	}

	data := map[string]interface{}{
		"data": pem,
	}
	_, err = ks.vault.Client.Logical().Write(filepath.Join(ks.path, ks.getPathForAlias(alias, "key")), data)
	if err != nil {
		logger.Errorf("Failed storing key [%s]: [%s]", alias, err)
		return err
	}

	return nil
}

func (ks *vaultBasedKeyStore) loadPrivateKey(alias string) (interface{}, error) {
	path := ks.getPathForAlias(alias, "sk")
	logger.Debugf("Loading private key [%s] at [%s]...", alias, path)

	raw, err := ks.vault.Client.Logical().Read(path)
	if err != nil {
		logger.Errorf("Failed loading private key [%s]: [%s].", alias, err.Error())
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(raw.Data["data"].(string))
	if err != nil {
		logger.Errorf("Failed to decode base64 private key [%s]: [%s].", alias, err.Error())
		return nil, err
	}
	privateKey, err := pemToPrivateKey(decoded, ks.pwd)
	if err != nil {
		logger.Errorf("Failed parsing private key [%s]: [%s].", alias, err.Error())
		return nil, err
	}

	return privateKey, nil
}

func (ks *vaultBasedKeyStore) loadPublicKey(alias string) (interface{}, error) {
	path := ks.getPathForAlias(alias, "pk")
	logger.Debugf("Loading public key [%s] at [%s]...", alias, path)

	raw, err := ks.vault.Client.Logical().Read(path)
	if err != nil {
		logger.Errorf("Failed loading public key [%s]: [%s].", alias, err.Error())

		return nil, err
	}

	privateKey, err := pemToPublicKey(raw.Data["data"].([]byte), ks.pwd)
	if err != nil {
		logger.Errorf("Failed parsing private key [%s]: [%s].", alias, err.Error())

		return nil, err
	}

	return privateKey, nil
}

func (ks *vaultBasedKeyStore) loadKey(alias string) ([]byte, error) {
	path := ks.getPathForAlias(alias, "key")
	logger.Debugf("Loading key [%s] at [%s]...", alias, path)

	pem, err := ks.vault.Client.Logical().Read(path)
	if err != nil {
		logger.Errorf("Failed loading key [%s]: [%s].", alias, err.Error())

		return nil, err
	}

	key, err := pemToAES(pem.Data["data"].([]byte), ks.pwd)
	if err != nil {
		logger.Errorf("Failed parsing key [%s]: [%s]", alias, err)

		return nil, err
	}

	return key, nil
}

func (ks *vaultBasedKeyStore) createKeyStore() error {
	return nil
}

func (ks *vaultBasedKeyStore) openKeyStore() error {
	if ks.isOpen {
		return nil
	}
	ks.isOpen = true
	logger.Debugf("KeyStore opened at [%s]...done", ks.path)

	return nil
}

func (ks *vaultBasedKeyStore) getPathForAlias(alias, suffix string) string {
	return filepath.Join(ks.path, alias+"_"+suffix)
}

func (ks *vaultBasedKeyStore) dirExists(path string) (bool, error) {
	raw, err := ks.vault.Client.Logical().Read(path)
	return raw != nil, err
}

func (ks *vaultBasedKeyStore) dirEmpty(path string) (bool, error) {
	raw, err := ks.vault.Client.Logical().List(path)
	return raw == nil, err
}
