/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package msp

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

var mspLogger = flogging.MustGetLogger("msp")

type mspManagerImpl struct {
	// map that contains all MSPs that we have setup or otherwise added
	mspsMap map[string]MSP

	// map that maps MSPs by their provider types
	mspsByProviders map[ProviderType][]MSP

	// error that might have occurred at startup
	up bool
}

// NewMSPManager returns a new MSP manager instance;
// note that this instance is not initialized until
// the Setup method is called
func NewMSPManager() MSPManager {
	return &mspManagerImpl{}
}

// Setup initializes the internal data structures of this manager and creates MSPs
func (mgr *mspManagerImpl) Setup(msps []MSP) error {
	if mgr.up {
		mspLogger.Infof("MSP manager already up")
		return nil
	}

	mspLogger.Debugf("Setting up the MSP manager (%d msps)", len(msps))

	// create the map that assigns MSP IDs to their manager instance - once
	mgr.mspsMap = make(map[string]MSP)

	// create the map that sorts MSPs by their provider types
	mgr.mspsByProviders = make(map[ProviderType][]MSP)

	for _, msp := range msps {
		// add the MSP to the map of active MSPs
		mspID, err := msp.GetIdentifier()
		if err != nil {
			return errors.WithMessage(err, "could not extract msp identifier")
		}
		mgr.mspsMap[mspID] = msp
		providerType := msp.GetType()
		mgr.mspsByProviders[providerType] = append(mgr.mspsByProviders[providerType], msp)
	}

	mgr.up = true

	mspLogger.Debugf("MSP manager setup complete, setup %d msps", len(msps))

	return nil
}

// GetMSPs returns the MSPs that are managed by this manager
func (mgr *mspManagerImpl) GetMSPs() (map[string]MSP, error) {
	return mgr.mspsMap, nil
}

// DeserializeIdentity returns an identity given its serialized version supplied as argument
func (mgr *mspManagerImpl) DeserializeIdentity(serializedID []byte) (Identity, error) {
	// We first deserialize to a SerializedIdentity to get the MSP ID
	sId := &msp.SerializedIdentity{}
	err := proto.Unmarshal(serializedID, sId)
	if err != nil {
		return nil, errors.Wrap(err, "could not deserialize a SerializedIdentity")
	}

	// we can now attempt to obtain the MSP
	msp := mgr.mspsMap[sId.Mspid]
	if msp == nil {
		return nil, errors.Errorf("MSP %s is unknown", sId.Mspid)
	}

	switch t := msp.(type) {
	case *bccspmsp:
		return t.deserializeIdentityInternal(sId.IdBytes)
	default:
		return t.DeserializeIdentity(serializedID)
	}
}

func (mgr *mspManagerImpl) IsWellFormed(identity *msp.SerializedIdentity) error {
	// Iterate over all the MSPs by their providers, and find at least 1 MSP that can attest
	// that this identity is well formed
	for _, mspList := range mgr.mspsByProviders {
		// We are guaranteed to have at least 1 MSP in each list from the initialization at Setup()
		msp := mspList[0]
		if err := msp.IsWellFormed(identity); err == nil {
			return nil
		}
	}
	return errors.New("no MSP provider recognizes the identity")
}
