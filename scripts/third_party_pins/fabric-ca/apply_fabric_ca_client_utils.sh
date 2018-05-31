#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the BCCSP package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/src/gofilter/cmd/gofilter/gofilter.go"

declare -a PKGS=(
    "api"
    "lib"
    "lib/streamer"
    "lib/tls"
    "lib/client"
    "lib/client/credential"
    "lib/client/credential/x509"
    "lib/common"
    "sdkpatch/logbridge"
    "sdkpatch/cryptosuitebridge"
    "util"
)

declare -a FILES=(
    "api/client.go"
    "api/net.go"

    "lib/client.go"
    "lib/identity.go"
    "lib/clientconfig.go"
    "lib/util.go"
    "lib/serverrevoke.go"
    "lib/sdkpatch_serverstruct.go"

    "lib/streamer/jsonstreamer.go"

    "lib/tls/tls.go"

    "lib/client/credential/credential.go"
    "lib/client/credential/x509/credential.go"
    "lib/client/credential/x509/signer.go"

    "lib/common/serverresponses.go"

    "sdkpatch/logbridge/logbridge.go"
    "sdkpatch/logbridge/syslogwriter.go"
    "sdkpatch/cryptosuitebridge/cryptosuitebridge.go"

    "util/util.go"
    "util/csp.go"
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}"
mkdir -p "${INTERNAL_PATH}"

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Apply fine-grained patching
gofilter() {
    echo "Filtering: ${FILTER_FILENAME}"
    cp ${TMP_PROJECT_PATH}/${FILTER_FILENAME} ${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak
    $GOFILTER_CMD -filename "${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak" \
        -filters "$FILTERS_ENABLED" -fn "$FILTER_FN" -gen "$FILTER_GEN" -mode "$FILTER_MODE" \
        > "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
} 

echo "Filtering Go sources for allowed functions ..."

FILTER_FILENAME="api/net.go"
START_LINE=`grep -n "IdemixEnrollmentRequestNet is" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..5}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done


FILTER_MODE="allow"
FILTERS_ENABLED="fn"

FILTER_FILENAME="lib/client.go"
FILTER_FN="Enroll,GenCSR,SendReq,Init,newPost,newEnrollmentResponse,newCertificateRequest,newPut,newGet,newDelete,StreamResponse"
FILTER_FN+=",getURL,NormalizeURL,initHTTPClient,net2LocalServerInfo,NewIdentity,newCfsslBasicKeyRequest"
FILTER_FN+=",handleIdemixEnroll,checkX509Enrollment,handleX509Enroll,GetCSP,NewX509Identity,net2LocalCAInfo"
gofilter
sed -i'' -e 's/util.GetServerPort()/\"\"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/log "github.com\// a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/\/\/ Initialize BCCSP (the crypto layer)/c.csp = cfg.CSP/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "c.csp, err = util.InitBCCSP(&cfg.CSP, mspDir, c.HomeDir)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..4}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
START_LINE=`grep -n "err := tls.AbsTLSClient(&c.Config.TLS, c.HomeDir)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..4}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/func NewCredential(certFile, keyFile string, c Client)/func NewCredential(keyFile core.Key, certFile []byte, c Client)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "func (c \*Client) handleIdemixEnroll(req" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
let "START_LINE+=1"
for i in {1..68}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/return c.newIdemixEnrollmentResponse(identity, &result, sk, req.Name)/return nil, errors.New("idemix enroll not supported")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/x509Cred := x509cred.NewCredential(c.certFile, c.keyFile, c)/x509Cred := x509cred.NewCredential(key, certByte, c)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"


FILTER_FILENAME="lib/identity.go"
FILTER_FN="newIdentity,Revoke,Post,addTokenAuthHdr,GetECert,Reenroll,Register,GetName,GetAllIdentities,GetIdentity,AddIdentity,ModifyIdentity,RemoveIdentity,Get,Put,Delete,GetStreamResponse,NewIdentity"
gofilter
sed -i'' -e 's/util.GetDefaultBCCSP()/nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/log "github.com\// a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="lib/clientconfig.go"
FILTER_FN=
gofilter
sed -i'' -e 's/*factory.FactoryOpts/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"


FILTER_FILENAME="lib/util.go"
FILTER_FN="GetCertID,BytesToX509Cert,addQueryParm"
gofilter

FILTER_FILENAME="lib/streamer/jsonstreamer.go"
FILTER_FN="StreamJSONArray,StreamJSON,stream,getNextName,skipToDelim,getSearchElement,getToken,errCB"
gofilter

FILTER_FILENAME="lib/tls/tls.go"
FILTER_FN="GetClientTLSConfig,checkCertDates"
gofilter
sed -i'' -e '/log "github.com\// a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "// ServerTLSConfig defines key material for a TLS server" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..14}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/CertFiles \[\]string `help:"A list of comma-separated PEM-encoded trusted certificate files (e.g. root1.pem,root2.pem)"`/CertFiles \[\]\[\]byte `help:"A list of comma-separated PEM-encoded trusted certificate bytes"`/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/KeyFile  string `help:"PEM-encoded key file when mutual authentication is enabled"`/KeyFile  []byte `help:"PEM-encoded key bytes when mutual authentication is enabled"`/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/CertFile string `help:"PEM-encoded certificate file when mutual authenticate is enabled"`/CertFile []byte `help:"PEM-encoded certificate bytes when mutual authenticate is enabled"`/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/\log.Debugf("Client Cert File:/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/\log.Debugf("Client Key File:/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/\log.Debugf("CA Files:/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/cfg.Client.CertFile != ""/cfg.Client.CertFile != nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "caCert, err := ioutil.ReadFile(cacert)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..4}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/caCert/cacert/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/errors.Errorf("Failed to process certificate from file %s", cacert)/errors.New("Failed to process certificate")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/func checkCertDates(certFile string) error {/func checkCertDates(certPEM []byte) error {/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "certPEM, err := ioutil.ReadFile(certFile)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..4}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done

FILTER_FILENAME="lib/client/credential/x509/credential.go"
FILTER_FN=",NewCredential,Type,Val,EnrollmentID,SetVal,Load,Store,CreateToken,RevokeSelf,getCSP"
gofilter
sed -i'' -e '/"encoding\/hex"/ a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/"github.com\/cloudflare/ a\
factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca\/sdkpatch\/cryptosuitebridge"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/return util.GetDefaultBCCSP()/return factory.GetDefault()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/certFile string/certFile []byte/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/keyFile  string/keyFile  core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/func NewCredential(certFile, keyFile string, c Client)/func NewCredential(keyFile core.Key, certFile []byte, c Client)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "func (cred \*Credential) Load() error {" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
let "START_LINE+=1"
for i in {1..15}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/cred.val, err = NewSigner(key, cert)/var err error \
    cred.val, err = NewSigner(cred.keyFile, cred.certFile)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "func (cred \*Credential) Store() error {" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
let "START_LINE+=1"
for i in {1..7}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/log.Infof("Stored client certificate at %s", cred.certFile)/log.Debugf("Credential.Store() not supported")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"




FILTER_FILENAME="lib/client/credential/x509/signer.go"
FILTER_FN=",NewSigner,Key,Cert,GetX509Cert,GetName,Attributes"
gofilter
sed -i'' -e '/"github.com\/cloudflare/ a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"


FILTER_FILENAME="util/csp.go"
FILTER_FN=",getBCCSPKeyOpts,ImportBCCSPKeyFromPEM,LoadX509KeyPair,GetSignerFromCert,BCCSPKeyRequestGenerate,GetSignerFromCertFile"
gofilter
sed -i'' -e '/_.\"time\"/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/\"github.com\/cloudflare\/cfssl\/cli\"/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/\"github.com\/cloudflare\/cfssl\/ocsp\"/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/log "github.com\// a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&factory.SwOpts{}/factory.NewSwOpts()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&factory.FileKeystoreOpts{}/factory.NewFileKeystoreOpts()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAKeyGenOpts{Temporary: ephemeral}/factory.GetECDSAKeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.RSA2048KeyGenOpts{Temporary: ephemeral}/factory.GetRSA2048KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.RSA3072KeyGenOpts{Temporary: ephemeral}/factory.GetRSA3072KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.RSA4096KeyGenOpts{Temporary: ephemeral}/factory.GetRSA4096KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}/factory.GetECDSAP256KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAP384KeyGenOpts{Temporary: ephemeral}/factory.GetECDSAP384KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAP512KeyGenOpts{Temporary: ephemeral}/factory.GetECDSAP512KeyGenOpts(ephemeral)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.X509PublicKeyImportOpts{Temporary: true}/factory.GetX509PublicKeyImportOpts(true)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAPrivateKeyImportOpts{Temporary: temporary}/factory.GetECDSAPrivateKeyImportOpts(temporary)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/cspsigner.New(/factory.NewCspSigner(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/utils.PrivateKeyToDER/factory.PrivateKeyToDER/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/utils.PEMtoPrivateKey/factory.PEMtoPrivateKey/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/key, err := factory.PEMtoPrivateKey(keyBuff, nil)/ i\
	key, err := ImportBCCSPKeyFromPEMBytes(keyBuff, myCSP, temporary) \
	if err != nil { \
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed parsing private key from key file %s", keyFile)) \
	} \
	return key, nil \
} \
\/\/ ImportBCCSPKeyFromPEMBytes attempts to create a private BCCSP key from a pem byte slice \
func ImportBCCSPKeyFromPEMBytes(keyBuff []byte, myCSP core.CryptoSuite, temporary bool) (core.Key, error) { \
keyFile := "pem bytes" \
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/LoadX509KeyPair(certFile, keyFile string, csp core.CryptoSuite)/LoadX509KeyPair(certFile, keyFile []byte, csp core.CryptoSuite)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/certPEMBlock, err := ioutil.ReadFile(certFile)/ i\
 certPEMBlock := certFile\
 ' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "certPEMBlock, err := ioutil.ReadFile(certFile)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..4}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e 's/errors.Errorf("Failed to find PEM block in file %s", certFile)/errors.New("Failed to find PEM block in bytes")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/errors.Errorf("Failed to find certificate PEM data in file %s, but did find a private key; PEM inputs may have been switched", certFile)/errors.New("Failed to find certificate PEM data in bytes, but did find a private key; PEM inputs may have been switched")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)/errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in bytes after skipping PEM blocks of the following types: %v", skippedBlockTypes)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/keyFile != ""/keyFile != nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/tls.LoadX509KeyPair(certFile, keyFile)/tls.X509KeyPair(certFile, keyFile)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/log.Debugf("Attempting fallback with certfile %s and keyfile %s", certFile, keyFile)/log.Debug("Attempting fallback with provided certfile and keyfile")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/return nil, errors.Wrapf(err, "Could not get the private key %s that matches %s", keyFile, certFile)/return nil, errors.Wrap(err, "Could not get the private key that matches the provided cert")/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="util/util.go"
FILTER_FN="ReadFile,HTTPRequestToString,HTTPResponseToString"
FILTER_FN+=",GetX509CertificateFromPEM,GetSerialAsHex,GetEnrollmentIDFromPEM"
FILTER_FN+=",MakeFileAbs,Marshal,StructToString,LoadX509KeyPair,CreateToken"
FILTER_FN+=",GenECDSAToken,GetEnrollmentIDFromX509Certificate,B64Encode,B64Decode"
FILTER_FN+=",GetMaskedURL,WriteFile,FileExists"
gofilter
sed -i'' -e '/log "golang.org\/x/ a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/mrand "math\// a\
factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca\/sdkpatch\/cryptosuitebridge"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.SHAOpts{}/factory.GetSHAOpts()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="lib/serverrevoke.go"
FILTER_FN=
gofilter

# Apply patching
echo "Patching import paths on upstream project ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" IMPORT_SUBSTS="${IMPORT_SUBSTS[@]}" scripts/third_party_pins/common/apply_import_patching.sh

echo "Inserting modification notice ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" scripts/third_party_pins/common/apply_header_notice.sh


# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done
