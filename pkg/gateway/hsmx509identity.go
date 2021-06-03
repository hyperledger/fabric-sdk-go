/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"encoding/json"
)

const Hsmx509type = "HSM-X.509"

// Hsmx509Identity represents an Hsmx509 identity
type Hsmx509Identity struct {
	IDType      string      `json:"type"`
	Version     int         `json:"version"`
	MspID       string      `json:"mspId"`
	Credentials credentials `json:"credentials"`
}

// Type returns Hsmx509 for this identity type
func (x *Hsmx509Identity) idType() string {
	return Hsmx509type
}

func (x *Hsmx509Identity) mspID() string {
	return x.MspID
}

// Certificate returns the Hsmx509 certificate PEM
func (x *Hsmx509Identity) Certificate() string {
	return x.Credentials.Certificate
}

// Key returns the private key PEM
func (x *Hsmx509Identity) Key() string {
	return x.Credentials.Key
}

// NewHsmx509Identity creates an Hsmx509 identity for storage in a wallet
func NewHsmx509Identity(mspid string, cert string, key string) *Hsmx509Identity {
	return &Hsmx509Identity{Hsmx509type, 1, mspid, credentials{cert, key}}
}

func (x *Hsmx509Identity) toJSON() ([]byte, error) {
	return json.Marshal(x)
}

func (x *Hsmx509Identity) fromJSON(data []byte) (Identity, error) {
	err := json.Unmarshal(data, x)

	if err != nil {
		return nil, err
	}

	return x, nil
}
