/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// Identity ...
type Identity struct {
	theType string
}

// IdentityType ...
type IdentityType interface {
	GetType() string
}

// IDHandler ...
type IDHandler interface {
	GetElements(id IdentityType) map[string]string
	FromElements(map[string]string) IdentityType
}
