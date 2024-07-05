package types

import (
	"context"
	"github.com/cockroachdb/apd/v3"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/internal/util"
	"time"
)

type GetRecordsInput struct {
	DateFrom *civil.Date
	DateTo   *civil.Date
	FrozenAt *time.Time
}

type StreamRecord struct {
	DateValue civil.Date
	Value     apd.Decimal
}

type IStream interface {
	AllowReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
	DisableReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
	AllowComposeStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error)
	DisableComposeStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error)
	GetComposeVisibility(ctx context.Context) (*util.VisibilityEnum, error)
	SetComposeVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
	GetReadVisibility(ctx context.Context) (*util.VisibilityEnum, error)
	GetAllowedReadWallets(ctx context.Context) ([]util.EthereumAddress, error)
	GetAllowedComposeStreams(ctx context.Context) ([]util.StreamId, error)
	SetReadVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
	InitializeStream(ctx context.Context) (transactions.TxHash, error)
	GetRecords(ctx context.Context, input GetRecordsInput) ([]StreamRecord, error)
}
