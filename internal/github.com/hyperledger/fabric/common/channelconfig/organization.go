/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package channelconfig

import (
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go/common"
	mspprotos "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
)

const (
	// MSPKey is the key for the MSP definition in orderer groups
	MSPKey = "MSP"
)

// OrganizationProtos are used to deserialize the organization config
type OrganizationProtos struct {
	MSP *mspprotos.MSPConfig
}

// OrganizationConfig stores the configuration for an organization
type OrganizationConfig struct {
	protos *OrganizationProtos

	mspConfigHandler *MSPConfigHandler
	msp              msp.MSP
	mspID            string
	name             string
}

// NewOrganizationConfig creates a new config for an organization
func NewOrganizationConfig(name string, orgGroup *cb.ConfigGroup, mspConfigHandler *MSPConfigHandler) (*OrganizationConfig, error) {
	if len(orgGroup.Groups) > 0 {
		return nil, fmt.Errorf("organizations do not support sub-groups")
	}

	oc := &OrganizationConfig{
		protos:           &OrganizationProtos{},
		name:             name,
		mspConfigHandler: mspConfigHandler,
	}

	if err := DeserializeProtoValuesFromGroup(orgGroup, oc.protos); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize values")
	}

	if err := oc.Validate(); err != nil {
		return nil, err
	}

	return oc, nil
}

// Name returns the name this org is referred to in config
func (oc *OrganizationConfig) Name() string {
	return oc.name
}

// MSPID returns the MSP ID associated with this org
func (oc *OrganizationConfig) MSPID() string {
	return oc.mspID
}

// Validate returns whether the configuration is valid
func (oc *OrganizationConfig) Validate() error {
	return oc.validateMSP()
}

func (oc *OrganizationConfig) validateMSP() error {
	var err error

	logger.Debugf("Setting up MSP for org %s", oc.name)
	oc.msp, err = oc.mspConfigHandler.ProposeMSP(oc.protos.MSP)
	if err != nil {
		return err
	}

	oc.mspID, _ = oc.msp.GetIdentifier()

	if oc.mspID == "" {
		return fmt.Errorf("MSP for org %s has empty MSP ID", oc.name)
	}

	return nil
}
