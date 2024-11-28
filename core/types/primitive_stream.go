package types

import (
	"context"

	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type InsertRecordInput struct {
	DateValue civil.Date
	Value     float64
}

type IPrimitiveStream interface {
	// IStream methods are also available in IPrimitiveStream
	IStream
	// InsertRecords inserts records into the stream
	InsertRecords(ctx context.Context, inputs []InsertRecordInput) (transactions.TxHash, error)
}
