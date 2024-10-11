package contractsapi

import (
	"context"
	"fmt"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
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

func LoadComposedStream(opts NewStreamOptions) (*ComposedStream, error) {
	stream, err := LoadStream(opts)
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
	StartDate string `json:"start_date"` // cannot use *string nor *civil.Date as decoding it will cause an error
}

func (c *ComposedStream) DescribeTaxonomies(ctx context.Context, params types.DescribeTaxonomiesParams) (types.Taxonomy, error) {
	records, err := c.call(ctx, "describe_taxonomies", []any{params.LatestVersion})

	if err != nil {
		return types.Taxonomy{}, err
	}

	result, err := DecodeCallResult[DescribeTaxonomiesResult](records)
	if err != nil {
		return types.Taxonomy{}, err
	}

	var taxonomyItems []types.TaxonomyItem
	for _, r := range result {
		dpAddress, err := util.NewEthereumAddressFromString(r.ChildDataProvider)
		if err != nil {
			return types.Taxonomy{}, err
		}
		weight, err := strconv.ParseFloat(r.Weight, 64)
		if err != nil {
			return types.Taxonomy{}, err
		}

		taxonomyItems = append(taxonomyItems, types.TaxonomyItem{
			ChildStream: types.StreamLocator{
				StreamId:     r.ChildStreamId,
				DataProvider: dpAddress,
			},
			Weight: weight,
		})
	}

	var startDateCivil *civil.Date
	if len(result) > 0 && result[0].StartDate != "" {
		parsedDate, err := civil.ParseDate(result[0].StartDate)
		if err != nil {
			return types.Taxonomy{}, err
		}
		startDateCivil = &parsedDate
	}

	return types.Taxonomy{
		TaxonomyItems: taxonomyItems,
		StartDate:     startDateCivil,
	}, nil
}

func (c *ComposedStream) SetTaxonomy(ctx context.Context, taxonomies types.Taxonomy) (transactions.TxHash, error) {
	var (
		dataProviders []string
		streamIDs     util.StreamIdSlice
		weights       []string
		startDate     string // null string is not able to be encoded by kwil, so lets left it empty by default
	)

	for _, taxonomy := range taxonomies.TaxonomyItems {
		dataProviderHexString := taxonomy.ChildStream.DataProvider.Address()
		// kwil expects no 0x prefix
		dataProviderHex := dataProviderHexString[2:]
		dataProviders = append(dataProviders, fmt.Sprintf("%s", dataProviderHex))
		streamIDs = append(streamIDs, taxonomy.ChildStream.StreamId)
		weights = append(weights, fmt.Sprintf("%f", taxonomy.Weight))
	}
	if taxonomies.StartDate != nil {
		startDate = taxonomies.StartDate.String()
	}

	var args [][]any
	args = append(args, []any{dataProviders, streamIDs.Strings(), weights, startDate})
	return c.checkedExecute(ctx, "set_taxonomy", args)
}
