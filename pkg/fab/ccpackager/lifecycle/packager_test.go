/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lifecycle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
)

const ccDir = "golang_cc"

func TestNewCCPackage(t *testing.T) {
	desc := &Descriptor{
		Path:  filepath.Join("./testdata", ccDir),
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: "example_cc",
	}

	pkgBytes, err := NewCCPackage(desc)
	require.NoError(t, err)

	metadataBytes, err := readMetadataFromBytes(pkgBytes)
	require.NoError(t, err)

	metadata := &PackageMetadata{}
	require.NoError(t, json.Unmarshal(metadataBytes, metadata))
	require.Equal(t, desc.Type.String(), metadata.Type)
	require.Equal(t, desc.Label, metadata.Label)
	require.Equalf(t, ccDir, metadata.Path, "expecting a normalized path")
}

func TestNewCCPackageError(t *testing.T) {
	t.Run("Empty path", func(t *testing.T) {
		desc := &Descriptor{
			Type:  pb.ChaincodeSpec_GOLANG,
			Label: "example_cc",
		}

		pkgBytes, err := NewCCPackage(desc)
		require.EqualError(t, err, "chaincode path must be specified")
		require.Empty(t, pkgBytes)
	})

	t.Run("Invalid path", func(t *testing.T) {
		desc := &Descriptor{
			Path:  "invalid",
			Type:  pb.ChaincodeSpec_GOLANG,
			Label: "example_cc",
		}

		pkgBytes, err := NewCCPackage(desc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "'go list' failed with: can't load package: package invalid is not in GOROOT")
		require.Empty(t, pkgBytes)
	})

	t.Run("Empty type", func(t *testing.T) {
		desc := &Descriptor{
			Path:  filepath.Join("./testdata", ccDir),
			Label: "example_cc",
		}

		pkgBytes, err := NewCCPackage(desc)
		require.EqualError(t, err, "chaincode language must be specified")
		require.Empty(t, pkgBytes)
	})

	t.Run("Empty label", func(t *testing.T) {
		desc := &Descriptor{
			Path: filepath.Join("./testdata", ccDir),
			Type: pb.ChaincodeSpec_GOLANG,
		}

		pkgBytes, err := NewCCPackage(desc)
		require.EqualError(t, err, "package label must be specified")
		require.Empty(t, pkgBytes)
	})

	t.Run("Invalid label", func(t *testing.T) {
		desc := &Descriptor{
			Path:  filepath.Join("./testdata", ccDir),
			Label: "this is chaincode example_cc",
			Type:  pb.ChaincodeSpec_GOLANG,
		}

		pkgBytes, err := NewCCPackage(desc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid label")
		require.Empty(t, pkgBytes)
	})
}

func TestGetTarGzBytesError(t *testing.T) {
	desc := &Descriptor{
		Path:  filepath.Join("./testdata", ccDir),
		Label: "example_cc",
		Type:  pb.ChaincodeSpec_GOLANG,
	}

	t.Run("Write metadata error", func(t *testing.T) {
		pkgBytes, err := getTarGzBytes(desc,
			func(tw *tar.Writer, name string, payload []byte) error {
				if name == metadataPackageName {
					return fmt.Errorf("metadata write error")
				}
				return nil
			},
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "metadata write error")
		require.Empty(t, pkgBytes)
	})

	t.Run("Write code package error", func(t *testing.T) {
		pkgBytes, err := getTarGzBytes(desc,
			func(tw *tar.Writer, name string, payload []byte) error {
				if name == codePackageName {
					return fmt.Errorf("code package write error")
				}
				return nil
			},
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "code package write error")
		require.Empty(t, pkgBytes)
	})

	t.Run("Normalize path error", func(t *testing.T) {
		desc := &Descriptor{
			Path:  filepath.Join("./testdata", ccDir),
			Label: "example_cc",
			Type:  pb.ChaincodeSpec_UNDEFINED,
		}

		pkgBytes, err := getTarGzBytes(desc, writePackage)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to normalize chaincode path")
		require.Empty(t, pkgBytes)
	})
}

func TestComputePackageID(t *testing.T) {
	packageID := ComputePackageID("label1", []byte("package"))
	require.NotEmpty(t, packageID)
}

func readMetadataFromBytes(pkgTarGzBytes []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(pkgTarGzBytes)
	gzr, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Name == "metadata.json" {
			return ioutil.ReadAll(tr)
		}
	}

	return nil, nil
}
