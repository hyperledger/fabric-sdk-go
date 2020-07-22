/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lifecycle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/chaincode/persistence"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/peer/packaging"
	"github.com/pkg/errors"
)

const (
	codePackageName     = "code.tar.gz"
	metadataPackageName = "metadata.json"
)

// NewCCPackage creates a chaincode package
func NewCCPackage(desc *Descriptor) ([]byte, error) {
	err := desc.Validate()
	if err != nil {
		return nil, err
	}

	pkgTarGzBytes, err := getTarGzBytes(desc, writePackage)
	if err != nil {
		return nil, err
	}

	return pkgTarGzBytes, nil
}

// Descriptor holds the package data
type Descriptor struct {
	Path  string
	Type  pb.ChaincodeSpec_Type
	Label string
}

// ComputePackageID returns the package ID from the given label and install package
func ComputePackageID(label string, pkgBytes []byte) string {
	return fmt.Sprintf("%s:%x", label, util.ComputeSHA256(pkgBytes))
}

// Validate validates the package descriptor
func (p *Descriptor) Validate() error {
	if p.Path == "" {
		return errors.New("chaincode path must be specified")
	}

	if p.Type == pb.ChaincodeSpec_UNDEFINED {
		return errors.New("chaincode language must be specified")
	}

	if p.Label == "" {
		return errors.New("package label must be specified")
	}

	if err := persistence.ValidateLabel(p.Label); err != nil {
		return err
	}

	return nil
}

// PackageMetadata holds the path and type for a chaincode package
type PackageMetadata struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type writer func(tw *tar.Writer, name string, payload []byte) error

func getTarGzBytes(desc *Descriptor, writeBytesToPackage writer) ([]byte, error) {
	payload := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(payload)
	tw := tar.NewWriter(gw)

	registry := packaging.NewRegistry(packaging.SupportedPlatforms...)

	normalizedPath, err := registry.NormalizePath(strings.ToUpper(desc.Type.String()), desc.Path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to normalize chaincode path")
	}
	metadataBytes, err := toJSON(normalizedPath, desc.Type.String(), desc.Label)
	if err != nil {
		return nil, err
	}
	err = writeBytesToPackage(tw, metadataPackageName, metadataBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package metadata to tar")
	}

	codeBytes, err := registry.GetDeploymentPayload(strings.ToUpper(desc.Type.String()), desc.Path)
	if err != nil {
		return nil, errors.WithMessage(err, "error getting chaincode bytes")
	}

	err = writeBytesToPackage(tw, codePackageName, codeBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package code bytes to tar")
	}

	err = tw.Close()
	if err == nil {
		err = gw.Close()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create tar for chaincode")
	}

	return payload.Bytes(), nil
}

func writePackage(tw *tar.Writer, name string, payload []byte) error {
	err := tw.WriteHeader(
		&tar.Header{
			Name: name,
			Size: int64(len(payload)),
			Mode: 0100644,
		},
	)
	if err != nil {
		return err
	}

	_, err = tw.Write(payload)
	return err
}

func toJSON(path, ccType, label string) ([]byte, error) {
	metadata := &PackageMetadata{
		Path:  path,
		Type:  ccType,
		Label: label,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode package metadata into JSON")
	}

	return metadataBytes, nil
}
