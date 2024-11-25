package types

import "github.com/trufnetwork/truf-node-sdk-go/core/util"

// StreamLocator is a struct that contains the StreamId and the DataProvider
type StreamLocator struct {
	// StreamId is the unique identifier of the stream, used as name of the deployed contract
	StreamId util.StreamId
	// DataProvider is the address of the data provider, it's the deployer of the stream
	DataProvider util.EthereumAddress
}
