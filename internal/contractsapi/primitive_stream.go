package contractsapi

import (
	"context"
	"fmt"
	"github.com/cockroachdb/apd/v3"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"strconv"
	"time"
)

/*
 * # PrimitiveStream
 * Represents the API interface to interact with a deployed Primitive stream.
 *
 * example:
 * - Insert data
 */

type PrimitiveStream struct {
	Stream
}

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

	if streamType != StreamTypePrimitive {
		return fmt.Errorf(ErrorStreamNotPrimitive)
	}

	return nil
}

func (p *PrimitiveStream) call(ctx context.Context, method string, args []any) (*client.CallResult, error) {
	err := p.checkValidPrimitiveStream(ctx)
	if err != nil {
		return nil, err
	}

	return p._client.Call(ctx, p.DBID, method, args)
}

func (p *PrimitiveStream) execute(ctx context.Context, method string, args [][]any) (transactions.TxHash, error) {
	err := p.checkValidPrimitiveStream(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	return p._client.Execute(ctx, p.DBID, method, args)
}

type InsertRecordInput struct {
	DateValue civil.Date
	Value     int
}

func (p *PrimitiveStream) InsertRecords(ctx context.Context, inputs []InsertRecordInput) (transactions.TxHash, error) {
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

	return p.execute(ctx, "insert_record", args)
}

type GetRecordOutput struct {
	DateValue civil.Date
	Value     apd.Decimal
}

type GetRecordRawOutput struct {
	DateValue string `json:"date_value"`
	Value     string `json:"value"`
}

type GetRecordsInput struct {
	DateFrom *civil.Date
	DateTo   *civil.Date
	FrozenAt *time.Time
}

// transformOrNil returns nil if the value is nil, otherwise it applies the transform function to the value.
func transformOrNil[T any](value *T, transform func(T) any) any {
	if value == nil {
		return nil
	}
	return transform(*value)
}

func (p *PrimitiveStream) GetRecords(ctx context.Context, input GetRecordsInput) ([]GetRecordOutput, error) {
	err := p.checkValidPrimitiveStream(ctx)
	if err != nil {
		return nil, err
	}

	var args []any
	args = append(args, transformOrNil(input.DateFrom, func(date civil.Date) any { return date.String() }))
	args = append(args, transformOrNil(input.DateTo, func(date civil.Date) any { return date.String() }))
	args = append(args, transformOrNil(input.FrozenAt, func(date time.Time) any { return date.UTC().Format(time.RFC3339) }))

	results, err := p.call(ctx, "get_record", args)
	if err != nil {
		return nil, err
	}

	rawOutputs, err := DecodeCallResult[GetRecordRawOutput](results)
	if err != nil {
		return nil, err
	}

	var outputs []GetRecordOutput
	for _, rawOutput := range rawOutputs {
		value, _, err := apd.NewFromString(rawOutput.Value)
		if err != nil {
			return nil, err
		}
		dateValue, err := civil.ParseDate(rawOutput.DateValue)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, GetRecordOutput{
			DateValue: dateValue,
			Value:     *value,
		})
	}

	return outputs, nil
}
