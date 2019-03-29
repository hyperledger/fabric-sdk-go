module github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos

replace github.com/hyperledger/fabric-sdk-go => ../../../../..

replace github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos => ./

require (
	github.com/golang/protobuf v1.2.0
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e // indirect
	google.golang.org/genproto v0.0.0-20190327125643-d831d65fe17d // indirect
	google.golang.org/grpc v1.19.0
)
