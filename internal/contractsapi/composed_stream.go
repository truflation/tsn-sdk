package contractsapi

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/internal/types"
	"github.com/truflation/tsn-sdk/internal/util"
	"strconv"
)

type ComposedStream struct {
	Stream
}

var _ types.IComposedStream = (*ComposedStream)(nil)

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

	if streamType != types.StreamTypeComposed {
		return fmt.Errorf(ErrorStreamNotComposed)
	}

	return nil
}

func (c *ComposedStream) checkedExecute(ctx context.Context, method string, args [][]any) (transactions.TxHash, error) {
	err := c.checkValidComposedStream(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	return c.execute(ctx, method, args)
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

func (c *ComposedStream) DescribeTaxonomies(ctx context.Context, params types.DescribeTaxonomiesParams) ([]types.TaxonomyItem, error) {
	records, err := c.call(ctx, "describe_taxonomies", []any{params.LatestVersion})

	if err != nil {
		return nil, err
	}

	result, err := DecodeCallResult[DescribeTaxonomiesResult](records)
	if err != nil {
		return nil, err
	}

	var taxonomies []types.TaxonomyItem
	for _, r := range result {
		dpAddress, err := util.NewEthereumAddressFromString(r.ChildDataProvider)
		if err != nil {
			return nil, err
		}
		weight, err := strconv.ParseFloat(r.Weight, 64)
		if err != nil {
			return nil, err
		}

		taxonomies = append(taxonomies, types.TaxonomyItem{
			ChildStream: types.StreamLocator{
				StreamId:     r.ChildStreamId,
				DataProvider: dpAddress,
			},
			Weight: weight,
		})
	}

	return taxonomies, nil
}

func (c *ComposedStream) SetTaxonomy(ctx context.Context, taxonomies []types.TaxonomyItem) (transactions.TxHash, error) {
	var (
		dataProviders []string
		streamIDs     util.StreamIdSlice
		weights       []string
	)

	for _, taxonomy := range taxonomies {
		dataProviders = append(dataProviders, fmt.Sprintf("%s", taxonomy.ChildStream.DataProvider.Address()))
		streamIDs = append(streamIDs, taxonomy.ChildStream.StreamId)
		weights = append(weights, fmt.Sprintf("%f", taxonomy.Weight))
	}

	var args [][]any

	args = append(args, []any{dataProviders, streamIDs.Strings(), weights})
	return c.checkedExecute(ctx, "set_taxonomy", args)
}
