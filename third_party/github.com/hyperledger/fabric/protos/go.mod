module github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos

replace github.com/hyperledger/fabric-sdk-go => ../../../../..

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos => ./

require (
	github.com/golang/protobuf v1.2.0
	github.com/hyperledger/fabric-sdk-go v0.0.0-20190125204638-b490519efff
)
