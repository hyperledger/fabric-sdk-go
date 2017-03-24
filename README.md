# Hyperledger Fabric Client SDK for Go

The Hyperledger Fabric Client SDK makes it easy to use APIs to interact with a Hyperledger Fabric blockchain.

This SDK is targeted both towards the external access to a Hyperledger Fabric blockchain using a Go application, as well as being targeted at the internal library in a peer to access API functions on other parts of the network.

**NOTE:** In an effort to make the codebase more modular, there will be interface changes over the course of the next week.

## Build and Test

This project must be cloned into `$GOPATH/src/github.com/hyperledger`. Package names have been chosen to match the Hyperledger project.

Execute `go test` from the project root to build the library and run the basic headless tests.

Execute `go test` in the `integration_test` to run end-to-end tests. This requires you to have:
- A working fabric and fabric-ca set up. Refer to the Hyperledger Fabric [documentation](https://github.com/hyperledger/fabric) on how to do this.
- Customized settings in the `integration_test/test_resources/config/config_test.yaml` in case your Hyperledger Fabric network is not running on `localhost` or is using different ports.

## Work in Progress

This client was last tested and found to be compatible with the following Hyperledger Fabric commit levels:
- fabric: `22d98b9e5ea36a6b209b3ea67def50a678718679`
- fabric-ca: `f18b6b769b80c889cb6b82ce34d755d9303ec881`

