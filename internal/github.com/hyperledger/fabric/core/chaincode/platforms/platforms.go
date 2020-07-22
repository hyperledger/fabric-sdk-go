/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package platforms

import (
	"archive/tar"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/chaincode/platforms/golang"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/chaincode/platforms/java"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/chaincode/platforms/node"
)

// SupportedPlatforms is the canonical list of platforms Fabric supports
var SupportedPlatforms = []Platform{
	&java.Platform{},
	&golang.Platform{},
	&node.Platform{},
}

// Interface for validating the specification and writing the package for
// the given platform
type Platform interface {
	Name() string
}

type PackageWriter interface {
	Write(name string, payload []byte, tw *tar.Writer) error
}

type PackageWriterWrapper func(name string, payload []byte, tw *tar.Writer) error

func (pw PackageWriterWrapper) Write(name string, payload []byte, tw *tar.Writer) error {
	return pw(name, payload, tw)
}

type Registry struct {
	Platforms     map[string]Platform
	PackageWriter PackageWriter
}

func NewRegistry(platformTypes ...Platform) *Registry {
	platforms := make(map[string]Platform)
	for _, platform := range platformTypes {
		if _, ok := platforms[platform.Name()]; ok {
			panic(fmt.Sprintf("Multiple platforms of the same name specified: %s", platform.Name()))
		}
		platforms[platform.Name()] = platform
	}
	return &Registry{
		Platforms:     platforms,
		PackageWriter: PackageWriterWrapper(writeBytesToPackage),
	}
}

func writeBytesToPackage(name string, payload []byte, tw *tar.Writer) error {
	err := tw.WriteHeader(&tar.Header{
		Name: name,
		Size: int64(len(payload)),
		Mode: 0100644,
	})
	if err != nil {
		return err
	}

	_, err = tw.Write(payload)
	if err != nil {
		return err
	}

	return nil
}
