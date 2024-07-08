package types

import "github.com/truflation/tsn-sdk/internal/util"

type StreamLocator struct {
	StreamId     util.StreamId
	DataProvider util.EthereumAddress
}
