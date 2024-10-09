package types

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type TaxonomyItem struct {
	ChildStream StreamLocator
	Weight      float64
	StartDate   string
}

type DescribeTaxonomiesParams struct {
	// LatestVersion if true, will return the latest version of the taxonomy only
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

// MarshalJSON Custom marshaler for TaxonomyDefinition
// TaxonomyDefinition -> ["st906974fb3f30a28200e907c604b15b",899]
func (t *TaxonomyItem) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{t.ChildStream.StreamId.String(), t.Weight})
}

// UnmarshalJSON Custom unmarshaller for TaxonomyDefinition
// ["st906974fb3f30a28200e907c604b15b",899] -> TaxonomyDefinition
func (t *TaxonomyItem) UnmarshalJSON(b []byte) error {
	var items []json.RawMessage
	err := json.Unmarshal(b, &items)
	if err != nil {
		return err
	}
	if len(items) != 2 {
		return fmt.Errorf("expected 2 elements, got %d", len(items))
	}

	// Unmarshal the first item as parentOf type
	if err := json.Unmarshal(items[0], &t.ChildStream.StreamId); err != nil {
		return fmt.Errorf("expected string, got error: %v", err)
	}

	// Unmarshal the second item as weight type
	if err := json.Unmarshal(items[1], &t.Weight); err != nil {
		return fmt.Errorf("expected float64, got error: %v", err)
	}

	return nil
}
