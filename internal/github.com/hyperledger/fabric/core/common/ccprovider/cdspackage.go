/*
Copyright IBM Corp. 2017 All Rights Reserved.

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

package ccprovider

import (
	"github.com/golang/protobuf/proto"
)

//----- CDSData ------

//CDSData is data stored in the LSCC on instantiation of a CC
//for CDSPackage.  This needs to be serialized for ChaincodeData
//hence the protobuf format
type CDSData struct {
	//CodeHash hash of CodePackage from ChaincodeDeploymentSpec
	CodeHash []byte `protobuf:"bytes,1,opt,name=codehash,proto3"`

	//MetaDataHash hash of Name and Version from ChaincodeDeploymentSpec
	MetaDataHash []byte `protobuf:"bytes,2,opt,name=metadatahash,proto3"`
}

//----implement functions needed from proto.Message for proto's mar/unmarshal functions

//Reset resets
func (data *CDSData) Reset() { *data = CDSData{} }

//String converts to string
func (data *CDSData) String() string { return proto.CompactTextString(data) }

//ProtoMessage just exists to make proto happy
func (*CDSData) ProtoMessage() {}
