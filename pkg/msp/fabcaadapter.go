/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/pkg/errors"

	"encoding/json"

	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib/client/credential"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib/client/credential/x509"
	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
)

// fabricCAAdapter translates between SDK lingo and native Fabric CA API
type fabricCAAdapter struct {
	config      msp.IdentityConfig
	cryptoSuite core.CryptoSuite
	caClient    *calib.Client
}

func newFabricCAAdapter(caID string, cryptoSuite core.CryptoSuite, config msp.IdentityConfig) (*fabricCAAdapter, error) {

	caClient, err := createFabricCAClient(caID, cryptoSuite, config)
	if err != nil {
		return nil, err
	}

	a := &fabricCAAdapter{
		config:      config,
		cryptoSuite: cryptoSuite,
		caClient:    caClient,
	}
	return a, nil
}

// Enroll handles enrollment.
func (c *fabricCAAdapter) Enroll(request *api.EnrollmentRequest) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]", request.Name)

	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName:  c.caClient.Config.CAName,
		Name:    request.Name,
		Secret:  request.Secret,
		Profile: request.Profile,
		Type:    request.Type,
		Label:   request.Label,
		CSR:     createCSRInfo(request.CSR),
	}

	if len(request.AttrReqs) > 0 {
		attrs := make([]*caapi.AttributeRequest, len(request.AttrReqs))
		for i, a := range request.AttrReqs {
			attrs[i] = &caapi.AttributeRequest{Name: a.Name, Optional: a.Optional}
		}
		careq.AttrReqs = attrs
	}

	caresp, err := c.caClient.Enroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Cert(), nil
}

// Reenroll handles re-enrollment
func (c *fabricCAAdapter) Reenroll(key core.Key, cert []byte, request *api.ReenrollmentRequest) ([]byte, error) {

	logger.Debugf("Re Enrolling user with provided key/cert pair for CA [%s]", c.caClient.Config.CAName)

	careq := &caapi.ReenrollmentRequest{
		CAName:  c.caClient.Config.CAName,
		Profile: request.Profile,
		Label:   request.Label,
		CSR:     createCSRInfo(request.CSR),
	}
	if len(request.AttrReqs) > 0 {
		attrs := make([]*caapi.AttributeRequest, len(request.AttrReqs))
		for i, a := range request.AttrReqs {
			attrs[i] = &caapi.AttributeRequest{Name: a.Name, Optional: a.Optional}
		}
		careq.AttrReqs = attrs
	}

	caidentity, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA signing identity")
	}

	caresp, err := caidentity.Reenroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "reenroll failed")
	}

	return caresp.Identity.GetECert().Cert(), nil
}

// Register handles user registration
// key: registrar private key
// cert: registrar enrollment certificate
// request: Registration Request
// Returns Enrolment Secret
func (c *fabricCAAdapter) Register(key core.Key, cert []byte, request *api.RegistrationRequest) (string, error) {
	// Construct request for Fabric CA client
	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}
	var req = caapi.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return "", errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke handles user revocation.
// key: registrar private key
// cert: registrar enrollment certificate
// request: Revocation Request
func (c *fabricCAAdapter) Revoke(key core.Key, cert []byte, request *api.RevocationRequest) (*api.RevocationResponse, error) {
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
		GenCRL: request.GenCRL,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	resp, err := registrar.Revoke(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	var revokedCerts []api.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			api.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	return &api.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

// GetCAInfo returns generic CA information
func (c *fabricCAAdapter) GetCAInfo(caname string) (*api.GetCAInfoResponse, error) {
	logger.Debugf("Get CA info [%s]", caname)

	req := &caapi.GetCAInfoRequest{CAName: caname}
	resp, err := c.caClient.GetCAInfo(req)
	if err != nil {
		return nil, errors.WithMessage(err, "GetCAInfo failed")
	}

	return getCAInfoResponse(resp), nil
}

func getCAInfoResponse(response *calib.GetCAInfoResponse) *api.GetCAInfoResponse {
	return &api.GetCAInfoResponse{
		CAName:                    response.CAName,
		CAChain:                   response.CAChain[:],
		IssuerPublicKey:           response.IssuerPublicKey[:],
		IssuerRevocationPublicKey: response.IssuerRevocationPublicKey[:],
		Version:                   response.Version,
	}
}

// CreateIdentity creates new identity
// key: registrar private key
// cert: registrar enrollment certificate
func (c *fabricCAAdapter) CreateIdentity(key core.Key, cert []byte, request *api.IdentityRequest) (*api.IdentityResponse, error) {

	logger.Debugf("Creating identity [%s:%s]", request.ID, request.Affiliation)

	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	// Create add identity request
	req := caapi.AddIdentityRequest{
		CAName:         request.CAName,
		ID:             request.ID,
		Affiliation:    request.Affiliation,
		Attributes:     attributes,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Secret:         request.Secret,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.AddIdentity(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add identity")
	}

	return getIdentityResponse(response), nil
}

// ModifyIdentity  modifies identity
// key: registrar private key
// cert: registrar enrollment certificate
func (c *fabricCAAdapter) ModifyIdentity(key core.Key, cert []byte, request *api.IdentityRequest) (*api.IdentityResponse, error) {

	logger.Debugf("Updating identity [%s:%s]", request.ID, request.Affiliation)

	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	// Create modify identity request
	req := caapi.ModifyIdentityRequest{
		CAName:         request.CAName,
		ID:             request.ID,
		Affiliation:    request.Affiliation,
		Attributes:     attributes,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Secret:         request.Secret,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.ModifyIdentity(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to modify identity")
	}

	return getIdentityResponse(response), nil
}

// RemoveIdentity  removes identity
// key: registrar private key
// cert: registrar enrollment certificate
func (c *fabricCAAdapter) RemoveIdentity(key core.Key, cert []byte, request *api.RemoveIdentityRequest) (*api.IdentityResponse, error) {

	logger.Debugf("Removing identity [%s]", request.ID)

	// Create remove request
	req := caapi.RemoveIdentityRequest{
		CAName: request.CAName,
		Force:  request.Force,
		ID:     request.ID,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.RemoveIdentity(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to remove identity")
	}

	return getIdentityResponse(response), nil
}

func getIdentityResponse(response *caapi.IdentityResponse) *api.IdentityResponse {

	var attributes []api.Attribute
	for i := range response.Attributes {
		attributes = append(attributes, api.Attribute{Name: response.Attributes[i].Name, Value: response.Attributes[i].Value, ECert: response.Attributes[i].ECert})
	}

	ret := &api.IdentityResponse{ID: response.ID,
		Affiliation:    response.Affiliation,
		Type:           response.Type,
		Attributes:     attributes,
		MaxEnrollments: response.MaxEnrollments,
		Secret:         response.Secret,
		CAName:         response.CAName,
	}

	return ret
}

// GetIdentity retrieves identity information
// key: registrar private key
// cert: registrar enrollment certificate
// id: identity id
func (c *fabricCAAdapter) GetIdentity(key core.Key, cert []byte, id, caname string) (*api.IdentityResponse, error) {

	logger.Debugf("Retrieving identity [%s]", id)

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.GetIdentity(id, caname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get identity")
	}

	var attributes []api.Attribute
	for i := range response.Attributes {
		attributes = append(attributes, api.Attribute{Name: response.Attributes[i].Name, Value: response.Attributes[i].Value, ECert: response.Attributes[i].ECert})
	}

	ret := &api.IdentityResponse{ID: response.ID,
		Affiliation:    response.Affiliation,
		Type:           response.Type,
		Attributes:     attributes,
		MaxEnrollments: response.MaxEnrollments,
		CAName:         response.CAName,
	}

	return ret, nil
}

// GetAllIdentities returns all identities that the caller is authorized to see
// key: registrar private key
// cert: registrar enrollment certificate
func (c *fabricCAAdapter) GetAllIdentities(key core.Key, cert []byte, caname string) ([]*api.IdentityResponse, error) {

	logger.Debug("Retrieving all identities")

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	var identities []caapi.IdentityInfo

	err = registrar.GetAllIdentities(caname, func(decoder *json.Decoder) error {
		var identity caapi.IdentityInfo
		decodeErr := decoder.Decode(&identity)
		if decodeErr != nil {
			return decodeErr
		}

		identities = append(identities, identity)
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get identities")
	}

	return getIdentityResponses(c.caClient.Config.CAName, identities), nil
}

// GetAffiliation returns information about the requested affiliation
func (c *fabricCAAdapter) GetAffiliation(key core.Key, cert []byte, affiliation, caname string) (*api.AffiliationResponse, error) {
	logger.Debugf("Retrieving affiliation [%s]", affiliation)

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	r, err := registrar.GetAffiliation(affiliation, caname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get affiliation")
	}

	resp := &api.AffiliationResponse{CAName: r.CAName, AffiliationInfo: api.AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// GetAllAffiliations returns all affiliations that the caller is authorized to see
func (c *fabricCAAdapter) GetAllAffiliations(key core.Key, cert []byte, caname string) (*api.AffiliationResponse, error) {
	logger.Debugf("Retrieving all affiliations")

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	r, err := registrar.GetAllAffiliations(caname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get affiliations")
	}

	resp := &api.AffiliationResponse{CAName: r.CAName, AffiliationInfo: api.AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// AddAffiliation add new affiliation
func (c *fabricCAAdapter) AddAffiliation(key core.Key, cert []byte, request *api.AffiliationRequest) (*api.AffiliationResponse, error) {
	logger.Debugf("Add affiliation [%s]", request.Name)

	req := caapi.AddAffiliationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Force:  request.Force,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	r, err := registrar.AddAffiliation(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add affiliation")
	}

	resp := &api.AffiliationResponse{CAName: r.CAName, AffiliationInfo: api.AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// ModifyAffiliation renames an existing affiliation on the server
func (c *fabricCAAdapter) ModifyAffiliation(key core.Key, cert []byte, request *api.ModifyAffiliationRequest) (*api.AffiliationResponse, error) {
	logger.Debugf("Updating affiliation [%s => %s]", request.Name, request.NewName)

	req := caapi.ModifyAffiliationRequest{
		CAName:  request.CAName,
		Name:    request.Name,
		NewName: request.NewName,
		Force:   request.Force,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	r, err := registrar.ModifyAffiliation(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to modify affiliation")
	}

	resp := &api.AffiliationResponse{CAName: r.CAName, AffiliationInfo: api.AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// RemoveAffiliation removes an existing affiliation from the server
func (c *fabricCAAdapter) RemoveAffiliation(key core.Key, cert []byte, request *api.AffiliationRequest) (*api.AffiliationResponse, error) {
	logger.Debugf("Removing affiliation [%s]", request.Name)

	// Create remove request
	req := caapi.RemoveAffiliationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Force:  request.Force,
	}

	registrar, err := c.newIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	r, err := registrar.RemoveAffiliation(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to remove affiliation")
	}

	resp := &api.AffiliationResponse{CAName: r.CAName, AffiliationInfo: api.AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

func fillAffiliationInfo(info *api.AffiliationInfo, name string, affiliations []caapi.AffiliationInfo, identities []caapi.IdentityInfo) error {
	info.Name = name

	// Add identities which have this affiliation
	idents := []api.IdentityInfo{}
	for _, identity := range identities {
		idents = append(idents, api.IdentityInfo{ID: identity.ID, Type: identity.Type, Affiliation: identity.Affiliation, Attributes: getAllAttributes(identity.Attributes), MaxEnrollments: identity.MaxEnrollments})
	}
	if len(idents) > 0 {
		info.Identities = idents
	}

	// Create child affiliations (if any)
	children := []api.AffiliationInfo{}
	for _, aff := range affiliations {
		childAff := api.AffiliationInfo{Name: aff.Name}
		err := fillAffiliationInfo(&childAff, aff.Name, aff.Affiliations, aff.Identities)
		if err != nil {
			return err
		}
		children = append(children, childAff)
	}
	if len(children) > 0 {
		info.Affiliations = children
	}
	return nil
}

func createCSRInfo(csr *api.CSRInfo) *caapi.CSRInfo {
	if csr == nil {
		// csr is not obrigatory, so we can return nil
		return nil
	}

	return &caapi.CSRInfo{
		CN:    csr.CN,
		Hosts: csr.Hosts,
	}
}

func getAllAttributes(attrs []caapi.Attribute) []api.Attribute {
	attriburtes := []api.Attribute{}
	for _, attr := range attrs {
		attriburtes = append(attriburtes, api.Attribute{Name: attr.Name, Value: attr.Value, ECert: attr.ECert})
	}

	return attriburtes
}

func (c *fabricCAAdapter) newIdentity(key core.Key, cert []byte) (*calib.Identity, error) {
	x509Cred := x509.NewCredential(key, cert, c.caClient)

	signer, err := x509.NewSigner(key, cert)
	if err != nil {
		return nil, err
	}

	err = x509Cred.SetVal(signer)
	if err != nil {
		return nil, err
	}

	return c.caClient.NewIdentity([]credential.Credential{x509Cred})
}

func getIdentityResponses(ca string, responses []caapi.IdentityInfo) []*api.IdentityResponse {

	ret := make([]*api.IdentityResponse, len(responses))

	for j, response := range responses {
		var attributes []api.Attribute
		for i := range response.Attributes {
			attributes = append(attributes, api.Attribute{Name: response.Attributes[i].Name, Value: response.Attributes[i].Value, ECert: response.Attributes[i].ECert})
		}
		ret[j] = &api.IdentityResponse{ID: response.ID,
			Affiliation:    response.Affiliation,
			Type:           response.Type,
			Attributes:     attributes,
			MaxEnrollments: response.MaxEnrollments,
			CAName:         ca,
		}
	}

	return ret
}

func createFabricCAClient(caID string, cryptoSuite core.CryptoSuite, config msp.IdentityConfig) (*calib.Client, error) {

	// Create new Fabric-ca client without configs
	c := &calib.Client{
		Config: &calib.ClientConfig{},
	}

	conf, ok := config.CAConfig(caID)
	if !ok {
		return nil, errors.Errorf("No CA '%s' in the configs", caID)
	}

	//set server CAName
	c.Config.CAName = conf.CAName
	//set server URL
	c.Config.URL = endpoint.ToAddress(conf.URL)
	//set server name
	c.Config.ServerName, _ = conf.GRPCOptions["ssl-target-name-override"].(string)
	//certs file list
	c.Config.TLS.CertFiles, ok = config.CAServerCerts(caID)
	if !ok {
		return nil, errors.Errorf("CA '%s' has no corresponding server certs in the configs", caID)
	}

	// set key file and cert file
	c.Config.TLS.Client.CertFile, ok = config.CAClientCert(caID)
	if !ok {
		return nil, errors.Errorf("CA '%s' has no corresponding client certs in the configs", caID)
	}

	c.Config.TLS.Client.KeyFile, ok = config.CAClientKey(caID)
	if !ok {
		return nil, errors.Errorf("CA '%s' has no corresponding client keys in the configs", caID)
	}

	var err error
	c.Config.TLS.TlsCertPool, err = config.TLSCACertPool().Get()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load configured cert pool")
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = endpoint.IsTLSEnabled(conf.URL)
	c.Config.MSPDir = config.CAKeyStorePath()

	//Factory opts
	c.Config.CSP = cryptoSuite

	err = c.Init()
	if err != nil {
		return nil, errors.Wrap(err, "CA Client init failed")
	}

	return c, nil
}
