package types

import (
	"context"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type TaxonomyItem struct {
	ChildStream StreamLocator
	Weight      float64
}

type DescribeTaxonomiesParams struct {
	LatestVersion bool
}

type IComposedStream interface {
	// IStream methods are also available in IPrimitiveStream
	IStream
	// DescribeTaxonomies returns the taxonomy of the stream
	DescribeTaxonomies(ctx context.Context, params DescribeTaxonomiesParams) ([]TaxonomyItem, error)
	// SetTaxonomy sets the taxonomy of the stream
	SetTaxonomy(ctx context.Context, taxonomies []TaxonomyItem) (transactions.TxHash, error)
}
