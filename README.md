# Hyperledger Fabric Client SDK for Go

The Hyperledger Fabric Client SDK makes it easy to use APIs to interact with a Hyperledger Fabric blockchain.

This SDK is targeted both towards the external access to a Hyperledger Fabric blockchain using a Go application, as well as being targeted at the internal library in a peer to access API functions on other parts of the network.

**NOTE:** In an effort to make the codebase more modular, there will be interface changes over the course of the next week.

## Build and Test

This project must be cloned into `$GOPATH/src/github.com/hyperledger`. Package names have been chosen to match the Hyperledger project.

Execute `go test` from the fabric-client and fabric-ca-client to build the library and run the basic headless tests.

Execute `go test` in the `test/integration` to run end-to-end tests. This requires you to have:
- A working fabric, fabric-ca and fabric-sdk-node set up. Refer to the Hyperledger Fabric [documentation](https://github.com/hyperledger/fabric) on how to do this.
- Customized settings in the `integration_test/test_resources/config/config_test.yaml` in case your Hyperledger Fabric network is not running on `localhost` or is using different ports.
- Run `create-channel.js` and `join-channel.js` in fabric-sdk-node test.


## Work in Progress

This client was last tested and found to be compatible with the following Hyperledger Fabric commit levels:
- fabric: v1.0.0-alpha 
- fabric-ca: `4651512e4e85728e6ecaf21b8cba52f51ed16633`
- fabric-sdk-node: v1.0.0-alpha