/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

// X509Identity represents an X509 identity
type X509Identity struct {
	Identity
	cert string
	key  string
}

// Type returns X509 for this identity type
func (x *X509Identity) Type() string {
	return "X509"
}

// Cert returns the X509 certificate PEM
func (x *X509Identity) Cert() string {
	return x.cert
}

// Key returns the private key PEM
func (x *X509Identity) Key() string {
	return x.key
}

// NewX509Identity creates an X509 identity for storage in a wallet
func NewX509Identity(cert string, key string) *X509Identity {
	return &X509Identity{Identity{"X509"}, cert, key}
}

type x509IdentityHandler struct{}

func (x *x509IdentityHandler) GetElements(id IdentityType) map[string]string {
	r, _ := id.(*X509Identity)

	return map[string]string{
		"cert": r.cert,
		"key":  r.key,
	}
}

func (x *x509IdentityHandler) FromElements(elements map[string]string) IdentityType {
	y := &X509Identity{Identity{"X509"}, elements["cert"], elements["key"]}
	return y
}

func newX509IdentityHandler() *x509IdentityHandler {
	return &x509IdentityHandler{}
}
