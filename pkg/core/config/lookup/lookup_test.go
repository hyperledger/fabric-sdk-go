/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lookup

import (
	"path/filepath"
	"testing"

	"os"

	"time"

	"strings"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var sampleConfigFile = filepath.Join("..", "testdata", "config_test_entity_matchers.yaml")

const orgChannelID = "orgchannel"

var backend *mocks.MockConfigBackend

type testEntityMatchers struct {
	matchers map[string][]MatchConfig
}

// networkConfig matches all network config elements
type networkConfig struct {
	Name                   string
	Description            string
	Version                string
	Client                 msp.ClientConfig
	Channels               map[string]fab.ChannelEndpointConfig
	Organizations          map[string]fab.OrganizationConfig
	Orderers               map[string]fab.OrdererConfig
	Peers                  map[string]fab.PeerConfig
	CertificateAuthorities map[string]msp.CAConfig
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames or channel names with static config
type MatchConfig struct {
	Pattern string

	// these are used for hostname mapping
	URLSubstitutionExp                  string
	SSLTargetOverrideURLSubstitutionExp string
	MappedHost                          string

	// this is used for Name mapping instead of hostname mappings
	MappedName string
}

func TestMain(m *testing.M) {
	backend = setupCustomBackend("key")
	r := m.Run()
	os.Exit(r)
}

func TestGetBool(t *testing.T) {
	//Test single backend lookup
	testLookup := New(backend)
	assert.True(t, testLookup.GetBool("key.bool.true"), "expected lookup to return true")
	assert.False(t, testLookup.GetBool("key.bool.false"), "expected lookup to return false")
	assert.False(t, testLookup.GetBool("key.bool.invalid"), "expected lookup to return false for invalid value")
	assert.False(t, testLookup.GetBool("key.bool.notexisting"), "expected lookup to return false for not existing value")

	//Test With multiple backend
	keyPrefixes := []string{"key1", "key2", "key3", "key4"}
	backends := getMultipleCustomBackends(keyPrefixes)
	testLookup = New(backends...)

	for _, prefix := range keyPrefixes {
		assert.True(t, testLookup.GetBool(prefix+".bool.true"), "expected lookup to return true")
		assert.False(t, testLookup.GetBool(prefix+".bool.false"), "expected lookup to return false")
		assert.False(t, testLookup.GetBool(prefix+".bool.invalid"), "expected lookup to return false for invalid value")
		assert.False(t, testLookup.GetBool(prefix+".bool.notexisting"), "expected lookup to return false for not existing value")
	}
}

func TestGetInt(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetInt("key.int.positive") == 5, "expected lookup to return valid positive value")
	assert.True(t, testLookup.GetInt("key.int.negative") == -5, "expected lookup to return valid negative value")
	assert.True(t, testLookup.GetInt("key.int.invalid") == 0, "expected lookup to return 0")
	assert.True(t, testLookup.GetInt("key.int.not.existing") == 0, "expected lookup to return 0")

	//Test With multiple backend
	keyPrefixes := []string{"key1", "key2", "key3", "key4"}
	backends := getMultipleCustomBackends(keyPrefixes)
	testLookup = New(backends...)

	for _, prefix := range keyPrefixes {
		assert.True(t, testLookup.GetInt(prefix+".int.positive") == 5, "expected lookup to return valid positive value")
		assert.True(t, testLookup.GetInt(prefix+".int.negative") == -5, "expected lookup to return valid negative value")
		assert.True(t, testLookup.GetInt(prefix+".int.invalid") == 0, "expected lookup to return 0")
		assert.True(t, testLookup.GetInt(prefix+".int.not.existing") == 0, "expected lookup to return 0")
	}
}

func TestGetString(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetString("key.string.valid") == "valid-string", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.lower.case") == "valid-string", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.upper.case") == "VALID-STRING", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.mixed.case") == "VaLiD-StRiNg", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.empty") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetString("key.string.nil") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetString("key.string.number") == "1234", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.not existing") == "", "expected lookup to return empty string value")

	//Test With multiple backend
	keyPrefixes := []string{"key1", "key2", "key3", "key4"}
	backends := getMultipleCustomBackends(keyPrefixes)
	testLookup = New(backends...)

	for _, prefix := range keyPrefixes {
		assert.True(t, testLookup.GetString(prefix+".string.valid") == "valid-string", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetString(prefix+".string.valid.lower.case") == "valid-string", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetString(prefix+".string.valid.upper.case") == "VALID-STRING", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetString(prefix+".string.valid.mixed.case") == "VaLiD-StRiNg", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetString(prefix+".string.empty") == "", "expected lookup to return empty string value")
		assert.True(t, testLookup.GetString(prefix+".string.nil") == "", "expected lookup to return empty string value")
		assert.True(t, testLookup.GetString(prefix+".string.number") == "1234", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetString(prefix+".string.not existing") == "", "expected lookup to return empty string value")
	}
}

func TestGetLowerString(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetLowerString("key.string.valid") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.lower.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.upper.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.mixed.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.empty") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetLowerString("key.string.nil") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetLowerString("key.string.number") == "1234", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetLowerString("key.string.not existing") == "", "expected lookup to return empty string value")

	//Test With multiple backends
	keyPrefixes := []string{"key1", "key2", "key3", "key4"}
	backends := getMultipleCustomBackends(keyPrefixes)
	testLookup = New(backends...)

	for _, prefix := range keyPrefixes {
		assert.True(t, testLookup.GetLowerString(prefix+".string.valid") == "valid-string", "expected lookup to return valid lowercase string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.valid.lower.case") == "valid-string", "expected lookup to return valid lowercase string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.valid.upper.case") == "valid-string", "expected lookup to return valid lowercase string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.valid.mixed.case") == "valid-string", "expected lookup to return valid lowercase string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.empty") == "", "expected lookup to return empty string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.nil") == "", "expected lookup to return empty string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.number") == "1234", "expected lookup to return valid string value")
		assert.True(t, testLookup.GetLowerString(prefix+".string.not existing") == "", "expected lookup to return empty string value")
	}
}

func TestGetDuration(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetDuration("key.duration.valid.hour").String() == (24*time.Hour).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.minute").String() == (24*time.Minute).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.second").String() == (24*time.Second).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.millisecond").String() == (24*time.Millisecond).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.microsecond").String() == (24*time.Microsecond).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.nanosecond").String() == (24*time.Nanosecond).String(), "expected valid time value")
	//default value tests
	assert.True(t, testLookup.GetDuration("key.duration.valid.not.existing").String() == (0*time.Second).String(), "expected valid default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.invalid").String() == (0*time.Second).String(), "expected valid  default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.nil").String() == (0*time.Second).String(), "expected valid  default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.empty").String() == (0*time.Second).String(), "expected valid  default time value")
	//default when no time unit provided
	assert.True(t, testLookup.GetDuration("key.duration.valid.no.unit").String() == (12*time.Nanosecond).String(), "expected valid default time value with default unit")

	//Test With multiple backends
	keyPrefixes := []string{"key1", "key2", "key3", "key4"}
	backends := getMultipleCustomBackends(keyPrefixes)
	testLookup = New(backends...)

	for _, prefix := range keyPrefixes {
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.hour").String() == (24*time.Hour).String(), "expected valid time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.minute").String() == (24*time.Minute).String(), "expected valid time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.second").String() == (24*time.Second).String(), "expected valid time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.millisecond").String() == (24*time.Millisecond).String(), "expected valid time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.microsecond").String() == (24*time.Microsecond).String(), "expected valid time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.nanosecond").String() == (24*time.Nanosecond).String(), "expected valid time value")
		//default value tests
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.not.existing").String() == (0*time.Second).String(), "expected valid default time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.invalid").String() == (0*time.Second).String(), "expected valid  default time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.nil").String() == (0*time.Second).String(), "expected valid  default time value")
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.empty").String() == (0*time.Second).String(), "expected valid  default time value")
		//default when no time unit provided
		assert.True(t, testLookup.GetDuration(prefix+".duration.valid.no.unit").String() == (12*time.Nanosecond).String(), "expected valid default time value with default unit")
	}
}

func TestUnmarshal(t *testing.T) {
	testLookup := New(backend)

	//output struct
	networkConfig := networkConfig{}
	testLookup.UnmarshalKey("channels", &networkConfig.Channels)

	assert.Equal(t, len(networkConfig.Channels), 6)
	assert.Equal(t, len(networkConfig.Channels["mychannel"].Peers), 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String(), (5 * time.Second).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String(), (500 * time.Millisecond).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor, 2.0)

	assert.Equal(t, fab.BlockHeightPriority, networkConfig.Channels["mychannel"].Policies.Selection.SortingStrategy)
	assert.Equal(t, fab.RoundRobin, networkConfig.Channels["mychannel"].Policies.Selection.Balancer)
	assert.Equal(t, 5, networkConfig.Channels["mychannel"].Policies.Selection.BlockHeightLagThreshold)
}

func TestUnmarshalWithMultipleBackend(t *testing.T) {

	sampleViper := newViper()

	var backends []core.ConfigBackend
	backendMap := make(map[string]interface{})
	backendMap["client"] = sampleViper.Get("client")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["channels"] = sampleViper.Get("channels")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["certificateAuthorities"] = sampleViper.Get("certificateAuthorities")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["entityMatchers"] = sampleViper.Get("entityMatchers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["organizations"] = sampleViper.Get("organizations")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["orderers"] = sampleViper.Get("orderers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["peers"] = sampleViper.Get("peers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	//create lookup with all 7 backends having 7 different entities
	testLookup := New(backends...)

	//output struct
	networkConfig := networkConfig{}
	entityMatchers := testEntityMatchers{}

	assert.Nil(t, testLookup.UnmarshalKey("client", &networkConfig.Client), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("channels", &networkConfig.Channels), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("entityMatchers", &entityMatchers.matchers), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("organizations", &networkConfig.Organizations), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("orderers", &networkConfig.Orderers), "unmarshalKey supposed to succeed")
	assert.Nil(t, testLookup.UnmarshalKey("peers", &networkConfig.Peers), "unmarshalKey supposed to succeed")

	//Client
	assert.True(t, networkConfig.Client.Organization == "org1")

	//Channel
	assert.Equal(t, len(networkConfig.Channels), 6)
	assert.Equal(t, len(networkConfig.Channels["mychannel"].Peers), 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets, 1)
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String(), (5 * time.Second).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String(), (500 * time.Millisecond).String())
	assert.Equal(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor, 2.0)

	//CertificateAuthorities
	assert.Equal(t, len(networkConfig.CertificateAuthorities), 2)
	assert.Equal(t, networkConfig.CertificateAuthorities["local.ca.org1.example.com"].URL, "https://ca.org1.example.com:7054")
	assert.Equal(t, networkConfig.CertificateAuthorities["local.ca.org2.example.com"].URL, "https://ca.org2.example.com:8054")

	//EntityMatchers
	assert.Equal(t, len(entityMatchers.matchers), 4)
	assert.Equal(t, len(entityMatchers.matchers["peer"]), 10)
	assert.Equal(t, entityMatchers.matchers["peer"][0].MappedHost, "local.peer0.org1.example.com")
	assert.Equal(t, len(entityMatchers.matchers["orderer"]), 6)
	assert.Equal(t, entityMatchers.matchers["orderer"][0].MappedHost, "local.orderer.example.com")
	assert.Equal(t, len(entityMatchers.matchers["certificateauthority"]), 3)
	assert.Equal(t, entityMatchers.matchers["certificateauthority"][0].MappedHost, "local.ca.org1.example.com")
	assert.Equal(t, len(entityMatchers.matchers["channel"]), 1)
	assert.Equal(t, entityMatchers.matchers["channel"][0].MappedName, "ch1")

	//Organizations
	assert.Equal(t, len(networkConfig.Organizations), 4)
	assert.Equal(t, networkConfig.Organizations["org1"].MSPID, "Org1MSP")

	//Orderer
	assert.Equal(t, len(networkConfig.Orderers), 3)
	assert.Equal(t, networkConfig.Orderers["local.orderer.example.com"].URL, "orderer.example.com:7050")

	//Peer
	assert.Equal(t, len(networkConfig.Peers), 4)
	assert.Equal(t, networkConfig.Peers["local.peer0.org1.example.com"].URL, "peer0.org1.example.com:7051")

}

func TestLookupUnmarshalAgainstViperUnmarshal(t *testing.T) {

	//new lookup
	testLookup := New(backend)
	//setup viper
	sampleViper := newViper()
	//viper network config
	networkConfigViper := networkConfig{}
	//lookup network config
	networkConfig := networkConfig{}

	/*
		TEST NETWORK CONFIG CLIENT
	*/
	//get client through backend lookup
	err := testLookup.UnmarshalKey("client", &networkConfig.Client)
	if err != nil {
		t.Fatal(err)
	}
	//get client from viper
	sampleViper.UnmarshalKey("client", &networkConfigViper.Client)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Client, &networkConfigViper.Client), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ORDERERS
	*/
	//get orderers through backend lookup
	err = testLookup.UnmarshalKey("orderers", &networkConfig.Orderers)
	if err != nil {
		t.Fatal(err)
	}
	//get orderers from viper
	sampleViper.UnmarshalKey("orderers", &networkConfigViper.Orderers)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Orderers, &networkConfigViper.Orderers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG CERTIFICATE AUTHORITIES
	*/
	//get certificate authorities through backend lookup
	err = testLookup.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	if err != nil {
		t.Fatal(err)
	}
	//get certificate authorities from viper
	sampleViper.UnmarshalKey("certificateAuthorities", &networkConfigViper.CertificateAuthorities)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.CertificateAuthorities, &networkConfigViper.CertificateAuthorities), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ENTITY MATCHERS
	*/
	entityMatchers := testEntityMatchers{}
	//get entityMatchers backend lookup
	err = testLookup.UnmarshalKey("entityMatchers", &entityMatchers.matchers)
	if err != nil {
		t.Fatal(err)
	}
	//get entityMatchers from viper
	viperEntityMatchers := testEntityMatchers{}
	sampleViper.UnmarshalKey("entityMatchers", &viperEntityMatchers.matchers)
	//now compare
	assert.True(t, reflect.DeepEqual(&entityMatchers, &viperEntityMatchers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ORGANIZATIONS
	*/
	//get organizations through backend lookup
	err = testLookup.UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		t.Fatal(err)
	}
	//get organizations from viper
	sampleViper.UnmarshalKey("organizations", &networkConfigViper.Organizations)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Organizations, &networkConfigViper.Organizations), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG CHANNELS
	*/
	//get channels through backend lookup
	err = testLookup.UnmarshalKey("channels", &networkConfig.Channels)
	if err != nil {
		t.Fatal(err)
	}
	//get channels from viper
	sampleViper.UnmarshalKey("channels", &networkConfigViper.Channels)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Channels, &networkConfigViper.Channels), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG PEERS
	*/
	//get peers through backend lookup
	err = testLookup.UnmarshalKey("peers", &networkConfig.Peers)
	if err != nil {
		t.Fatal(err)
	}
	//get peers from viper
	sampleViper.UnmarshalKey("peers", &networkConfigViper.Peers)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Peers, &networkConfigViper.Peers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	//Just to make sure that empty values are not being compared
	assert.True(t, len(networkConfigViper.Channels) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(viperEntityMatchers.matchers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.Orderers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.Peers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(entityMatchers.matchers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, networkConfigViper.Client.Organization != "", "expected to get valid unmarshalled value")

}

func setupCustomBackend(keyPrefix string) *mocks.MockConfigBackend {

	backendMap := make(map[string]interface{})

	backendMap[keyPrefix+".bool.true"] = true
	backendMap[keyPrefix+".bool.false"] = false
	backendMap[keyPrefix+".bool.invalid"] = "INVALID"

	backendMap[keyPrefix+".int.positive"] = 5
	backendMap[keyPrefix+".int.negative"] = -5
	backendMap[keyPrefix+".int.invalid"] = "INVALID"

	backendMap[keyPrefix+".string.valid"] = "valid-string"
	backendMap[keyPrefix+".string.valid.mixed.case"] = "VaLiD-StRiNg"
	backendMap[keyPrefix+".string.valid.lower.case"] = "valid-string"
	backendMap[keyPrefix+".string.valid.upper.case"] = "VALID-STRING"
	backendMap[keyPrefix+".string.empty"] = ""
	backendMap[keyPrefix+".string.nil"] = nil
	backendMap[keyPrefix+".string.number"] = 1234

	backendMap[keyPrefix+".duration.valid.hour"] = "24h"
	backendMap[keyPrefix+".duration.valid.minute"] = "24m"
	backendMap[keyPrefix+".duration.valid.second"] = "24s"
	backendMap[keyPrefix+".duration.valid.millisecond"] = "24ms"
	backendMap[keyPrefix+".duration.valid.microsecond"] = "24µs"
	backendMap[keyPrefix+".duration.valid.nanosecond"] = "24ns"
	backendMap[keyPrefix+".duration.valid.no.unit"] = "12"
	backendMap[keyPrefix+".duration.invalid"] = "24XYZ"
	backendMap[keyPrefix+".duration.nil"] = nil
	backendMap[keyPrefix+".duration.empty"] = ""

	//test fab network config
	sampleViper := newViper()
	backendMap["client"] = sampleViper.Get("client")
	backendMap["channels"] = sampleViper.Get("channels")
	backendMap["certificateAuthorities"] = sampleViper.Get("certificateAuthorities")
	backendMap["entityMatchers"] = sampleViper.Get("entityMatchers")
	backendMap["organizations"] = sampleViper.Get("organizations")
	backendMap["orderers"] = sampleViper.Get("orderers")
	backendMap["peers"] = sampleViper.Get("peers")

	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}

func getMultipleCustomBackends(keyPrefixes []string) []core.ConfigBackend {
	var backends []core.ConfigBackend
	for _, prefix := range keyPrefixes {
		backends = append(backends, setupCustomBackend(prefix))
	}
	return backends
}

func TestUnmarshalWithHookFunc(t *testing.T) {
	testLookup := New(backend)
	tamperPeerChannelConfig(backend)
	//output struct
	networkConfig := networkConfig{}
	testLookup.UnmarshalKey("channels", &networkConfig.Channels, WithUnmarshalHookFunction(setTrueDefaultForPeerChannelConfig()))

	//Test if mandatory hook func is working as expected
	assert.True(t, len(networkConfig.Channels) == 6)
	assert.True(t, len(networkConfig.Channels["mychannel"].Peers) == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String() == (5*time.Second).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String() == (500*time.Millisecond).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor == 2.0)

	//Test if custom hook func is working
	assert.True(t, len(networkConfig.Channels[orgChannelID].Peers) == 2)
	//test orgchannel peer1 (EndorsingPeer should be true as set, remaining should be default = true)
	orgChannelPeer1 := networkConfig.Channels[orgChannelID].Peers["peer0.org1.example.com"]
	assert.True(t, orgChannelPeer1.EndorsingPeer)
	assert.True(t, orgChannelPeer1.LedgerQuery)
	assert.True(t, orgChannelPeer1.EventSource)
	assert.True(t, orgChannelPeer1.ChaincodeQuery)

	//test orgchannel peer1 (EndorsingPeer should be false as set, remaining should be default = true)
	orgChannelPeer2 := networkConfig.Channels[orgChannelID].Peers["peer0.org2.example.com"]
	assert.False(t, orgChannelPeer2.EndorsingPeer)
	assert.True(t, orgChannelPeer2.LedgerQuery)
	assert.True(t, orgChannelPeer2.EventSource)
	assert.True(t, orgChannelPeer2.ChaincodeQuery)
}

func newViper() *viper.Viper {
	myViper := viper.New()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	myViper.SetConfigFile(sampleConfigFile)
	err := myViper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	return myViper
}

func tamperPeerChannelConfig(backend *mocks.MockConfigBackend) {
	channelsMap := backend.KeyValueMap["channels"]
	orgChannel := map[string]interface{}{
		"orderers": []string{"orderer.example.com"},
		"peers": map[string]interface{}{
			"peer0.org1.example.com": map[string]interface{}{"endorsingpeer": true},
			"peer0.org2.example.com": map[string]interface{}{"endorsingpeer": false},
		},
	}
	(channelsMap.(map[string]interface{}))[orgChannelID] = orgChannel
}

func setTrueDefaultForPeerChannelConfig() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		//If target is of type 'fab.PeerChannelConfig', then only hook should work
		if t == reflect.TypeOf(fab.PeerChannelConfig{}) {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				setDefault(dataMap, "endorsingpeer", true)
				setDefault(dataMap, "chaincodequery", true)
				setDefault(dataMap, "ledgerquery", true)
				setDefault(dataMap, "eventsource", true)

				return dataMap, nil
			}
		}
		return data, nil
	}
}

//setDefault sets default value provided to map if given key not found
func setDefault(dataMap map[string]interface{}, key string, defaultVal bool) {
	_, ok := dataMap[key]
	if !ok {
		dataMap[key] = true
	}
}
