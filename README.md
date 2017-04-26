# Hyperledger Fabric Client SDK for Go
[![Build Status](https://jenkins.hyperledger.org/buildStatus/icon?job=fabric-sdk-go-tests-merge-x86_64)](https://jenkins.hyperledger.org/job/fabric-sdk-go-tests-merge-x86_64)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger/fabric-sdk-go)](https://goreportcard.com/report/github.com/hyperledger/fabric-sdk-go)
[![GoDoc](https://godoc.org/github.com/hyperledger/fabric-sdk-go?status.svg)](https://godoc.org/github.com/hyperledger/fabric-sdk-go)

The Hyperledger Fabric Client SDK makes it easy to use APIs to interact with a Hyperledger Fabric blockchain.

This SDK is targeted both towards the external access to a Hyperledger Fabric blockchain using a Go application, as well as being targeted at the internal library in a peer to access API functions on other parts of the network.

This is a **read-only mirror** of the formal [Gerrit](https://gerrit.hyperledger.org/r/#/admin/projects/fabric-sdk-go)
repository, where active development is ongoing. Issue tracking is handled in [Jira](https://jira.hyperledger.org/secure/RapidBoard.jspa?projectKey=FAB&rapidView=7&view=planning)

## Build and Test

You need:
- A working fabric, and fabric-ca set up. It is recommended that you use the docker-compose file provided in `test/fixtures`. See steps below.
- Customized settings in the `test/fixtures/config/config_test.yaml` in case your Hyperledger Fabric network is not running on `localhost` or is using different ports.
```
# Build fabric:
cd $GOPATH/src/github.com/hyperledger/
git clone https://github.com/hyperledger/fabric
cd $GOPATH/src/github.com/hyperledger/fabric/
git checkout v1.0.0-alpha
make docker

# Build fabric-ca:
cd $GOPATH/src/github.com/hyperledger/
git clone https://github.com/hyperledger/fabric-ca
cd $GOPATH/src/github.com/hyperledger/fabric-ca/
git checkout v1.0.0-alpha
make docker

# Before running the test, make sure you don't have stale invalid certificates from previous runs
rm -rf /tmp/keystore/
rm -rf /tmp/enroll_user/

# Start fabric
cd $GOPATH/src/github.com/hyperledger/
git clone https://github.com/hyperledger/fabric-sdk-go
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/
docker-compose -f docker-compose.yaml up --force-recreate
```

Fabric should now be running. In a diferent shell, run integration tests
```
cd $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/integration/
go test
```

## Compatibility

This client was last tested and found to be compatible with the following Hyperledger Fabric commit levels:
- fabric: v1.0.0-alpha
- fabric-ca: v1.0.0-alpha
