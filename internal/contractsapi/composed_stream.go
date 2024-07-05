package contractsapi

import (
	"context"
	"fmt"
	"github.com/truflation/tsn-sdk/internal/util"
)

/*
 * # ComposedStream
 * Represents the API interface to interact with a deployed composed stream.
 * Example:
 * - Describe Taxonomies
 * - Insert Taxonomy
 */

type ComposedStream struct {
	Stream
}

const (
	ErrorStreamNotComposed = "stream is not a composed stream"
)

func ComposedStreamFromStream(ctx context.Context, stream Stream) (*ComposedStream, error) {
	streamType, err := stream.GetType(ctx)

	if err != nil {
		return nil, err
	}

	if streamType != StreamTypeComposed {
		return nil, fmt.Errorf(ErrorStreamNotComposed)
	}
	return &ComposedStream{
		Stream: stream,
	}, nil
}

func NewComposedStream(ctx context.Context, opts NewStreamOptions) (*ComposedStream, error) {
	stream, err := NewStream(opts)
	if err != nil {
		return nil, err
	}

	return ComposedStreamFromStream(ctx, *stream)
}

type DescribeTaxonomiesParams struct {
	LatestVersion bool
}

type DescribeTaxonomiesResult struct {
	ChildStreamId     util.StreamId `json:"child_stream_id"`
	ChildDataProvider string        `json:"child_data_provider"`
	// decimals are received as strings by kwil to avoid precision loss
	// as decimal are more arbitrary than golang's float64
	Weight    string `json:"weight"`
	CreatedAt int    `json:"created_at"`
	Version   int    `json:"version"`
}

func (s ComposedStream) DescribeTaxonomies(ctx context.Context, params DescribeTaxonomiesParams) ([]DescribeTaxonomiesResult, error) {
	records, err := s._client.Call(ctx, s.DBID, "describe_taxonomies", []any{params.LatestVersion})

	if err != nil {
		return nil, err
	}

	return DecodeCallResult[DescribeTaxonomiesResult](records)
}
