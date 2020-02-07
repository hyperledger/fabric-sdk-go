/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

// client options collector
type clientOptions struct {
	orgName string
	caID    string
}

// request options collector
type requestOptions struct {
	caName string
}

// enrollment options collector
type enrollmentOptions struct {
	secret   string
	profile  string
	label    string
	typ      string
	attrReqs []*AttributeRequest
	csr      *CSRInfo
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*clientOptions) error

// WithOrg option
func WithOrg(orgName string) ClientOption {
	return func(o *clientOptions) error {
		o.orgName = orgName
		return nil
	}
}

// WithCAInstance option
func WithCAInstance(caID string) ClientOption {
	return func(o *clientOptions) error {
		o.caID = caID
		return nil
	}
}

// RequestOption func for each Opts argument
type RequestOption func(*requestOptions) error

// WithCA allows for specifying optional CA name (within the CA server instance)
func WithCA(caName string) RequestOption {
	return func(o *requestOptions) error {
		o.caName = caName
		return nil
	}
}

// EnrollmentOption describes a functional parameter for Enroll
type EnrollmentOption func(*enrollmentOptions) error

// WithSecret enrollment option
func WithSecret(secret string) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.secret = secret
		return nil
	}
}

// WithProfile enrollment option
func WithProfile(profile string) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.profile = profile
		return nil
	}
}

// WithType enrollment option
func WithType(typ string) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.typ = typ
		return nil
	}
}

// WithLabel enrollment option
func WithLabel(label string) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.label = label
		return nil
	}
}

// WithAttributeRequests enrollment option
func WithAttributeRequests(attrReqs []*AttributeRequest) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.attrReqs = attrReqs
		return nil
	}
}

// WithCSR enrollment option
func WithCSR(csr *CSRInfo) EnrollmentOption {
	return func(o *enrollmentOptions) error {
		o.csr = csr
		return nil
	}
}
