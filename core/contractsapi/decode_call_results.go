package contractsapi

import (
	"encoding/json"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/client"
)

// DecodeCallResult decodes the result of a view call to the specified struct.
func DecodeCallResult[T any](result *client.Records) ([]T, error) {
	// Export returns all of the records in a slice. The map in each slice is
	// equivalent to a Record, which is keyed by the column name.
	records := result.Export()

	// Convert the []map[string]any to JSON bytes
	recordsBytes, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal records: %v", err)
	}

	// Unmarshal JSON bytes into a slice of getMetadataResult
	var results []T
	err = json.Unmarshal(recordsBytes, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %v", err)
	}

	return results, nil
}
