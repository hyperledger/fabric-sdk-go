# Hyperledger Fabric Client SDK for Go

[![Build Status](https://jenkins.hyperledger.org/buildStatus/icon?job=fabric-sdk-go-tests-merge-x86_64)](https://jenkins.hyperledger.org/job/fabric-sdk-go-tests-merge-x86_64)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger/fabric-sdk-go)](https://goreportcard.com/report/github.com/hyperledger/fabric-sdk-go)
[![GoDoc](https://godoc.org/github.com/hyperledger/fabric-sdk-go?status.svg)](https://godoc.org/github.com/hyperledger/fabric-sdk-go)

This SDK enables Go developers to build solutions that interact with [Hyperledger Fabric](http://hyperledger-fabric.readthedocs.io/en/latest/).

## Getting started

Obtain the client SDK packages for Fabric and Fabric CA.

```bash
go get -u github.com/hyperledger/fabric-sdk-go

# Optional - populate vendor directory (if needed by your downstream vendoring solution)
# cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/
# make populate
```

You're good to go, happy coding! Check out the examples for usage demonstrations.

### Examples

- [E2E Test](test/integration/e2e/end_to_end.go): Basic example that uses SDK to query and execute transaction
- [Ledger Query Test](test/integration/sdk/ledger_queries_test.go): Basic example that uses SDK to query a channel's underlying ledger
- [Multi Org Test](test/integration/orgs/multiple_orgs_test.go): An example that has multiple organisations involved in transaction
- [Dynamic Endorser Selection](test/integration/sdk/sdk_provider_test.go): An example that uses dynamic endorser selection (based on chaincode policy)
- [E2E PKCS11 Test](test/integration/pkcs11/e2e_test.go): E2E Test using a PKCS11 crypto suite and configuration
- [CLI](https://github.com/securekey/fabric-examples/tree/master/fabric-cli/): An example CLI for Fabric built with the Go SDK.
- More examples needed!

### Community

- Discussion is happening in [Rocket Chat](https://chat.hyperledger.org/channel/fabric-sdk-go).
- Issue tracking is handled in [Jira](https://jira.hyperledger.org/secure/RapidBoard.jspa?projectKey=FAB&rapidView=7&view=planning).
- Active development occurs in the [Gerrit](https://gerrit.hyperledger.org/r/#/admin/projects/fabric-sdk-go) repository.

## Client SDK

### Current Compatibility
The SDK's integration tests run against three tagged Fabric versions:
- prev (currently v1.0.6)
- stable (currently v1.1.0)
- prerelease (currently disabled)

Additionally for development purposes integration tests also run against the devstable Fabric version as needed.

### v1.0 Notes
The SDK uses the Fabric v1.1 delivery service as the default event mechanism. When using Fabric v1.0, you must override this default by setting the event service type in the config:

```
client:
  peer:
    eventService:
      type: eventhub
```

### Retired versions
When the 'prev' code level is updated, the last tested fabric-sdk-go commit or tag is listed below.

- fabric v1.0.0 & fabric-ca v1.0.0
  - fabric-sdk-go: 79b343ba

### Running the test suite

```bash
# In the Fabric SDK Go directory
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/

# Optional - Automatically install Go tools used by test suite
# make depend-install

# Running test suite
make

# Clean test suite run artifacts
make clean
```

### Go Tags
The following Go tags can be supplied to enable additional functionality:
- experimental: includes support for experimental features.

## Contributing to the Go SDK

If you want to contribute to the Go SDK, please run the test suite and submit patches to the Gerrit git repostory for review. For general guidelines, please refer to the Fabric project's [contribution page](http://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING.html).

You need:

- Go 1.9
- [Dep](https://github.com/golang/dep)
- Make
- Docker
- Docker Compose
- Git

### Gerrit Git repository

To contribute patches, you will need to clone (or add a remote) from [Gerrit](https://gerrit.hyperledger.org/r/#/admin/projects/fabric-sdk-go) with authentication.

### Running a portion of the test suite

```bash
# In the Fabric SDK Go directory
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/

# Ensure dependencies are installed
make depend

# Running code checks (license, linting, spelling, etc)
make checks

# Running all unit tests
make unit-test

# Running all integration tests
make integration-test
```

### Running package unit tests manually

```bash
# In a package directory
go test
```

### Running integration tests manually

You need:

- A working fabric and fabric-ca set up. It is recommended that you use the docker-compose file provided in `test/fixtures/dockerenv`. It is also recommended that you use the default .env settings provided in `test/fixtures/dockerenv`. See steps below.
- Customized settings in the `test/fixtures/config/config_test.yaml` in case your Hyperledger Fabric network is not running on `localhost` or is using different ports.

#### Testing with Fabric Images at Docker Hub

The test suite defaults to the latest compatible tag of fabric images at Docker Hub.
The following commands starts Fabric:

```bash
# In the Fabric SDK Go directory
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/

# Start fabric (stable tag)
make dockerenv-stable-up

# Or more generally, start fabric at a different code level (prev, stable, prerelease, devstable)
# make dockerenv-[CODELEVEL]-up
```

#### Running Integration Tests

Fabric should now be running. In a different shell, run integration tests

```bash
# In the Fabric SDK Go directory
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go

# Use script to setup parameters for integration tests and execute them
# Previously we use to have hostnames like Fabric CA server, orderer and peer pointed to localhost
# Now since we removed this now, We will be using a different configuration
make integration-tests-local

# Or more generally, run integration tests at a different code level (prev, stable, prerelease, devstable)
# and fixture target version
# FABRIC_CODELEVEL_VER=[VER] FABRIC_CODELEVEL_TAG=[CODELEVEL] make integration-tests-local
```


```bash
# Previously we use to have hostnames like Fabric CA server, orderer and peer pointed to localhost
# Now since we removed this now, We will be using a different config file config_test_local.yaml
# which has the Fabric CA server, orderer and peers pointed to localhost
# It is also possible to run integration tests using go test directly. For example:
#cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/integration/
#go -args testLocal=true test

#cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/integration/orgs
#go -args testLocal=true test

# You should review test/scripts/integration.sh for options and details.
# Note: you should generally prefer the scripted version to setup parameters for you.
```

#### Testing with Local Build of Fabric (Advanced)

Alternatively you can use a local build of Fabric using the following commands:

```bash
# Start fabric (devstable codelevel with latest docker tags)
make dockerenv-latest-up
```

## License

Hyperledger Fabric SDK Go software is licensed under the [Apache License Version 2.0](LICENSE).

---
This document is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.