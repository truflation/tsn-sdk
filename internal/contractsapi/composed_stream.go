package contractsapi

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/internal/util"
)

type ComposedStream struct {
	Stream
}

const (
	ErrorStreamNotComposed = "stream is not a composed stream"
)

func ComposedStreamFromStream(stream Stream) (*ComposedStream, error) {
	return &ComposedStream{
		Stream: stream,
	}, nil
}

func NewComposedStream(opts NewStreamOptions) (*ComposedStream, error) {
	stream, err := NewStream(opts)
	if err != nil {
		return nil, err
	}

	return ComposedStreamFromStream(*stream)
}

// checkValidComposedStream checks if the stream is a valid composed stream
// and returns an error if it is not. Valid means:
// - the stream is initialized
// - the stream is a composed stream
func (c *ComposedStream) checkValidComposedStream(ctx context.Context) error {
	// first check if is initialized
	err := c.checkInitialized(ctx)
	if err != nil {
		return err
	}

	// then check if is composed
	streamType, err := c.GetType(ctx)
	if err != nil {
		return err
	}

	if streamType != StreamTypeComposed {
		return fmt.Errorf(ErrorStreamNotComposed)
	}

	return nil
}

func (c *ComposedStream) call(ctx context.Context, method string, args []any) (*client.CallResult, error) {
	err := c.checkValidComposedStream(ctx)
	if err != nil {
		return nil, err
	}

	return c._client.Call(ctx, c.DBID, method, args)
}

func (c *ComposedStream) execute(ctx context.Context, method string, args [][]any) (transactions.TxHash, error) {
	err := c.checkValidComposedStream(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	return c._client.Execute(ctx, c.DBID, method, args)
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

func (c *ComposedStream) DescribeTaxonomies(ctx context.Context, params DescribeTaxonomiesParams) ([]DescribeTaxonomiesResult, error) {
	records, err := c.call(ctx, "describe_taxonomies", []any{params.LatestVersion})

	if err != nil {
		return nil, err
	}

	return DecodeCallResult[DescribeTaxonomiesResult](records)
}
