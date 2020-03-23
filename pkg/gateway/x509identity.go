/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import "encoding/json"

const x509Type = "X.509"

// X509Identity represents an X509 identity
type X509Identity struct {
	Version     int         `json:"version"`
	MspID       string      `json:"mspId"`
	IDType      string      `json:"type"`
	Credentials credentials `json:"credentials"`
}

type credentials struct {
	Certificate string `json:"certificate"`
	Key         string `json:"privateKey"`
}

// Type returns X509 for this identity type
func (x *X509Identity) idType() string {
	return x509Type
}

func (x *X509Identity) mspID() string {
	return x.MspID
}

// Certificate returns the X509 certificate PEM
func (x *X509Identity) Certificate() string {
	return x.Credentials.Certificate
}

// Key returns the private key PEM
func (x *X509Identity) Key() string {
	return x.Credentials.Key
}

// NewX509Identity creates an X509 identity for storage in a wallet
func NewX509Identity(mspid string, cert string, key string) *X509Identity {
	return &X509Identity{1, mspid, x509Type, credentials{cert, key}}
}

func (x *X509Identity) toJSON() ([]byte, error) {
	return json.Marshal(x)
}

func (x *X509Identity) fromJSON(data []byte) (Identity, error) {
	err := json.Unmarshal(data, x)

	if err != nil {
		return nil, err
	}

	return x, nil
}
