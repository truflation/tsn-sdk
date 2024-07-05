package types

import (
	"context"
	kwilClientPkg "github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	tsnapi "github.com/truflation/tsn-sdk/internal/contractsapi"
	"github.com/truflation/tsn-sdk/internal/util"
	"time"
)

type Client interface {
	WaitForTx(ctx context.Context, txHash transactions.TxHash, interval time.Duration) (*transactions.TcTxQueryResponse, error)
	KwilClient() *kwilClientPkg.Client
	DeployStream(ctx context.Context, streamId util.StreamId, streamType tsnapi.StreamType) (transactions.TxHash, error)
	DestroyStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error)
	LoadStream(streamId util.StreamId) (*tsnapi.Stream, error)
	LoadPrimitiveStream(streamId util.StreamId) (*tsnapi.PrimitiveStream, error)
	LoadComposedStream(streamId util.StreamId) (*tsnapi.ComposedStream, error)
}
