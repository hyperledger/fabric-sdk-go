/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package verifier provides various verifier (e.g. signature)
package verifier

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

const loggerModule = "fabsdk/client"

var logger = logging.NewLogger(loggerModule)

// Signature verifies response signature
type Signature struct {
	Membership fab.ChannelMembership
}

// Verify checks transaction proposal response
func (v *Signature) Verify(response *fab.TransactionProposalResponse) error {

	if response.ProposalResponse.GetResponse().Status < int32(common.Status_SUCCESS) || response.ProposalResponse.GetResponse().Status >= int32(common.Status_BAD_REQUEST) {
		return status.NewFromProposalResponse(response.ProposalResponse, response.Endorser)
	}

	res := response.ProposalResponse

	if res.GetEndorsement() == nil {
		return errors.WithStack(status.New(status.EndorserClientStatus, status.MissingEndorsement.ToInt32(), "missing endorsement in proposal response", nil))
	}
	creatorID := res.GetEndorsement().Endorser

	err := v.Membership.Validate(creatorID)
	if err != nil {
		return errors.WithStack(status.New(status.EndorserClientStatus, status.SignatureVerificationFailed.ToInt32(), "the creator certificate is not valid", []interface{}{err.Error()}))
	}

	// check the signature against the endorser and payload hash
	digest := append(res.GetPayload(), res.GetEndorsement().Endorser...)

	// validate the signature
	err = v.Membership.Verify(creatorID, digest, res.GetEndorsement().Signature)
	if err != nil {
		return errors.WithStack(status.New(status.EndorserClientStatus, status.SignatureVerificationFailed.ToInt32(), "the creator's signature over the proposal is not valid", []interface{}{err.Error()}))
	}

	return nil
}

// Match matches transaction proposal responses (empty for signature verifier)
func (v *Signature) Match(response []*fab.TransactionProposalResponse) error {
	return nil
}

//ValidateCertificateDates used to verify if certificate was expired or not valid until later date
func ValidateCertificateDates(cert *x509.Certificate) error {
	if cert == nil {
		return nil
	}
	if time.Now().UTC().Before(cert.NotBefore) {
		return errors.New("Certificate provided is not valid until later date")
	}

	if time.Now().UTC().After(cert.NotAfter) {
		return errors.New("Certificate provided has expired")
	}
	return nil
}

//VerifyPeerCertificate verifies raw certs and chain certs for expiry and not yet valid dates
func VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	for _, chaincert := range rawCerts {
		cert, err := x509.ParseCertificate(chaincert)
		if err != nil {
			logger.Warn("Got error while verifying cert")
		}
		if cert != nil {
			err = ValidateCertificateDates(cert)
			if err != nil {
				//cert is expired or not valid
				logger.Warn(err.Error())
				return err
			}
		}
	}
	for _, certs := range verifiedChains {
		for _, cert := range certs {
			err := ValidateCertificateDates(cert)
			if err != nil {
				//cert is expired or not valid
				logger.Warn(err.Error())
				return err
			}
		}
	}
	return nil
}
