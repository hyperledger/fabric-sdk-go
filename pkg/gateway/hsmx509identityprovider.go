package gateway

import (
	"encoding/json"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

const ProviderType = "HSM-X.509"


type HsmOptions struct {
	Lib         string      `json:"lib"`
	Pin         string      `json:"pin"`
	Slot        string      `json:"slot"`
	UserType    string      `json:"usertype"`
	ReadWrite   bool      `json:"readwrite"`
}

type Hsmx509IdentityProvider struct {
	Type           string             `json:"type"`
	Options        HsmOptions         `json:"options"`
	CryptoSuite    core.CryptoSuite   `json:"cryptosuite"`
}

/*
func (p Hsmx509IdentityProvider) GetCryptoSuite() core.CryptoSuite {

}
*/

func (p Hsmx509IdentityProvider) FromJson(data []byte) (Identity, error) {
	err := json.Unmarshal(data, p)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p Hsmx509IdentityProvider) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

/*
func (p Hsmx509IdentityProvider) GetUserContext(identity Hsmx509Identity, name string) (Identity, error) {
	return
}
*/