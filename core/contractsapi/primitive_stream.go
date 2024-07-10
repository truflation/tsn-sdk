package contractsapi

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/core/types"
	"strconv"
)

type PrimitiveStream struct {
	Stream
}

var _ types.IPrimitiveStream = (*PrimitiveStream)(nil)

const (
	ErrorStreamNotPrimitive = "stream is not a primitive stream"
)

func PrimitiveStreamFromStream(stream Stream) (*PrimitiveStream, error) {
	return &PrimitiveStream{
		Stream: stream,
	}, nil
}

func NewPrimitiveStream(options NewStreamOptions) (*PrimitiveStream, error) {
	stream, err := NewStream(options)
	if err != nil {
		return nil, err
	}
	return PrimitiveStreamFromStream(*stream)
}

// checkValidPrimitiveStream checks if the stream is a valid primitive stream
// and returns an error if it is not. Valid means:
// - the stream is initialized
// - the stream is a primitive stream
func (p *PrimitiveStream) checkValidPrimitiveStream(ctx context.Context) error {
	// first check if is initialized
	err := p.checkInitialized(ctx)
	if err != nil {
		return err
	}

	// then check if is primitive
	streamType, err := p.GetType(ctx)
	if err != nil {
		return err
	}

	if streamType != types.StreamTypePrimitive {
		return fmt.Errorf(ErrorStreamNotPrimitive)
	}

	return nil
}

func (p *PrimitiveStream) checkedExecute(ctx context.Context, method string, args [][]any) (transactions.TxHash, error) {
	err := p.checkValidPrimitiveStream(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	return p._client.Execute(ctx, p.DBID, method, args)
}

func (p *PrimitiveStream) InsertRecords(ctx context.Context, inputs []types.InsertRecordInput) (transactions.TxHash, error) {
	err := p.checkValidPrimitiveStream(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	var args [][]any
	for _, input := range inputs {

		dateStr := input.DateValue.String()

		args = append(args, []any{
			dateStr,
			strconv.Itoa(input.Value),
		})
	}

	return p.checkedExecute(ctx, "insert_record", args)
}
