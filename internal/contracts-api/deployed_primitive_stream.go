package tsn_api

import (
	"context"
	"fmt"
	apd "github.com/cockroachdb/apd/v3"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"strconv"
	"time"
)

/*
 * # DeployedPrimitiveStream
 * Represents the API interface to interact with a deployed Primitive stream.
 *
 * example:
 * - Insert data
 */

type DeployedPrimitiveStream struct {
	DeployedStream
}

const (
	ErrorStreamNotPrimitive = "stream is not a primitive stream"
)

func DeployedPrimitiveStreamFromDeployedStream(ctx context.Context, stream DeployedStream) (*DeployedPrimitiveStream, error) {
	streamType, err := stream.GetType(ctx)

	if err != nil {
		return nil, err
	}

	if streamType != StreamTypePrimitive {
		return nil, fmt.Errorf(ErrorStreamNotPrimitive)
	}
	return &DeployedPrimitiveStream{
		DeployedStream: stream,
	}, nil
}

func NewDeployedPrimitiveStream(ctx context.Context, options NewDeployedStreamOptions) (*DeployedPrimitiveStream, error) {
	stream, err := NewDeployedStream(options)
	if err != nil {
		return nil, err
	}
	return DeployedPrimitiveStreamFromDeployedStream(ctx, *stream)
}

type InsertRecordInput struct {
	DateValue civil.Date
	Value     int
}

func (s *DeployedPrimitiveStream) InsertRecords(ctx context.Context, inputs []InsertRecordInput) (transactions.TxHash, error) {
	var args [][]any
	for _, input := range inputs {

		dateStr := input.DateValue.String()

		args = append(args, []any{
			dateStr,
			strconv.Itoa(input.Value),
		})
	}

	return s._client.Execute(ctx, s.DBID, "insert_record", args)
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

func (s *DeployedPrimitiveStream) GetRecords(ctx context.Context, input GetRecordsInput) ([]GetRecordOutput, error) {
	var args []any
	args = append(args, transformOrNil(input.DateFrom, func(date civil.Date) any { return date.String() }))
	args = append(args, transformOrNil(input.DateTo, func(date civil.Date) any { return date.String() }))
	args = append(args, transformOrNil(input.FrozenAt, func(date time.Time) any { return date.UTC().Format(time.RFC3339) }))

	results, err := s._client.Call(ctx, s.DBID, "get_record", args)
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
