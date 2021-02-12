package sw

import (
	"bytes"
	"crypto/ecdsa"
	"github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"net/http"
	"path/filepath"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Vault struct {
	Client *api.Client
	space  string
}

func NewVault(addr, token, space string) *Vault {
	client, err := api.NewClient(&api.Config{Address: addr, HttpClient: httpClient})
	if err != nil {
		panic(err)
	}

	client.SetToken(token)
	return &Vault{Client: client, space: space}
}

func (vault *Vault) RecursiveSearch(vaultPath string, ski []byte, pwd []byte) (bccsp.Key, error) {
	data, err := vault.Client.Logical().List(vaultPath)
	if err != nil {
		return nil, err
	}
	if data != nil {
		for _, key := range data.Data["keys"].([]interface{}) {
			_, err := vault.RecursiveSearch(filepath.Join(vaultPath, key.(string)), ski, pwd)
			if err != nil {
				return nil, err
			}
		}
	} else {
		data, err := vault.Client.Logical().Read(vaultPath)
		if data == nil {
			return nil, nil
		}
		key, err := pemToPrivateKey([]byte(data.Data["data"].(string)), pwd)
		if err != nil {
			return nil, err
		}

		var k *ecdsaPrivateKey
		switch kk := key.(type) {
		case *ecdsa.PrivateKey:
			k = &ecdsaPrivateKey{kk, true}
		default:
			return nil, nil
		}

		if !bytes.Equal(k.SKI(), ski) {
			return nil, nil
		}

		return k, nil
	}

	return nil, nil
}
