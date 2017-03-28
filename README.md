# Hyperledger Fabric Client SDK for Go
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger/fabric-sdk-go)](https://goreportcard.com/report/github.com/hyperledger/fabric-sdk-go)
[![GoDoc](https://godoc.org/github.com/hyperledger/fabric-sdk-go?status.svg)](https://godoc.org/github.com/hyperledger/fabric-sdk-go)

The Hyperledger Fabric Client SDK makes it easy to use APIs to interact with a Hyperledger Fabric blockchain.

This SDK is targeted both towards the external access to a Hyperledger Fabric blockchain using a Go application, as well as being targeted at the internal library in a peer to access API functions on other parts of the network.

**NOTE:** In an effort to make the codebase more modular, there will be interface changes over the course of the next week.

This is a **read-only mirror** of the formal [Gerrit](https://gerrit.hyperledger.org/r/#/admin/projects/fabric-sdk-go)
repository, where active development is ongoing. Issue tracking is handled in [Jira](https://jira.hyperledger.org/secure/RapidBoard.jspa?projectKey=FAB&rapidView=7&view=planning)

## Build and Test

This project must be cloned into `$GOPATH/src/github.com/hyperledger`. Package names have been chosen to match the Hyperledger project.

Execute `go test` from the fabric-client and fabric-ca-client to build the library and run the basic headless tests.

Execute `go test` in the `test/integration` to run end-to-end tests. This requires you to have:
- A working fabric, fabric-ca and fabric-sdk-node set up. Refer to the Hyperledger Fabric [documentation](https://github.com/hyperledger/fabric) on how to do this.
- Customized settings in the `integration_test/test_resources/config/config_test.yaml` in case your Hyperledger Fabric network is not running on `localhost` or is using different ports.
- Run `create-channel.js` and `join-channel.js` in fabric-sdk-node test.


## Compatibility

This client was last tested and found to be compatible with the following Hyperledger Fabric commit levels:
- fabric: v1.0.0-alpha 
- fabric-ca: `4651512e4e85728e6ecaf21b8cba52f51ed16633`
- fabric-sdk-node: v1.0.0-alpha