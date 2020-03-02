/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// Identity represents a specific identity format
type Identity interface {
	idType() string
	mspID() string
	toJSON() ([]byte, error)
	fromJSON(data []byte) (Identity, error)
}
