package types

import (
	"context"
	"github.com/cockroachdb/apd/v3"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/core/util"
	"time"
)

type GetRecordInput struct {
	DateFrom *civil.Date
	DateTo   *civil.Date
	FrozenAt *time.Time
}

type GetIndexInput = GetRecordInput

type StreamRecord struct {
	DateValue civil.Date
	Value     apd.Decimal
}

type StreamIndex = StreamRecord

type IStream interface {
	// InitializeStream initializes the stream. Majority of other methods need the stream to be initialized
	InitializeStream(ctx context.Context) (transactions.TxHash, error)
	// GetRecord reads the records of the stream within the given date range
	GetRecord(ctx context.Context, input GetRecordInput) ([]StreamRecord, error)
	// GetIndex reads the index of the stream within the given date range
	GetIndex(ctx context.Context, input GetIndexInput) ([]StreamIndex, error)
	// GetType gets the type of the stream -- Primitive or Composed
	GetType(ctx context.Context) (StreamType, error)

	// SetReadVisibility sets the read visibility of the stream -- Private or Public
	SetReadVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
	// GetReadVisibility gets the read visibility of the stream -- Private or Public
	GetReadVisibility(ctx context.Context) (*util.VisibilityEnum, error)
	// SetComposeVisibility sets the compose visibility of the stream -- Private or Public
	SetComposeVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
	// GetComposeVisibility gets the compose visibility of the stream -- Private or Public
	GetComposeVisibility(ctx context.Context) (*util.VisibilityEnum, error)

	// AllowReadWallet allows a wallet to read the stream, if reading is private
	AllowReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
	// DisableReadWallet disables a wallet from reading the stream
	DisableReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
	// AllowComposeStream allows a stream to use this stream as child, if composing is private
	AllowComposeStream(ctx context.Context, locator StreamLocator) (transactions.TxHash, error)
	// DisableComposeStream disables a stream from using this stream as child
	DisableComposeStream(ctx context.Context, locator StreamLocator) (transactions.TxHash, error)

	// GetAllowedReadWallets gets the wallets allowed to read the stream
	GetAllowedReadWallets(ctx context.Context) ([]util.EthereumAddress, error)
	// GetAllowedComposeStreams gets the streams allowed to compose this stream
	GetAllowedComposeStreams(ctx context.Context) ([]StreamLocator, error)
}
