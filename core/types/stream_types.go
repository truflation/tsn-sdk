package types

import "github.com/truflation/tsn-sdk/core/util"

type StreamLocator struct {
	StreamId     util.StreamId
	DataProvider util.EthereumAddress
}
