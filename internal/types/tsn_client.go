package types

import (
	"context"
	kwilClientPkg "github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/internal/util"
	"time"
)

type Client interface {
	WaitForTx(ctx context.Context, txHash transactions.TxHash, interval time.Duration) (*transactions.TcTxQueryResponse, error)
	KwilClient() *kwilClientPkg.Client
	DeployStream(ctx context.Context, streamId util.StreamId, streamType StreamType) (transactions.TxHash, error)
	DestroyStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error)
	LoadStream(stream StreamLocator) (IStream, error)
	LoadPrimitiveStream(stream StreamLocator) (IPrimitiveStream, error)
	LoadComposedStream(stream StreamLocator) (IComposedStream, error)
	/*
	 * utils for the client
	 */
	OwnStreamLocator(streamId util.StreamId) StreamLocator
	Address() util.EthereumAddress
}
