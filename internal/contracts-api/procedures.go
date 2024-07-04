package tsn_api

import (
	"context"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/pkg/errors"
	"reflect"
)

// ## View only procedures

type GetMetadataParams struct {
	Key        MetadataKey
	OnlyLatest bool
	// optional. Gets metadata with ref value equal to the given value
	Ref string
}

type GetMetadataResult struct {
	RowId  string `json:"row_id"`
	ValueI int    `json:"value_i"`
	ValueB bool   `json:"value_b"`
	// TODO: uncomment when supported
	// ValueF    float64 `json:"value_f"`
	ValueS    string `json:"value_s"`
	ValueRef  string `json:"value_ref"`
	CreatedAt int    `json:"created_at"`
}

// GetValueByKey returns the value of the metadata by its key
// I.e. if we expect an int from `ComposeVisibility`, we can call this function
// to get `valueI` from the result
func (g GetMetadataResult) GetValueByKey(t MetadataKey) (any, error) {
	metadataType := t.GetType()

	switch metadataType {
	case MetadataTypeInt:
		return g.ValueI, nil
	case MetadataTypeBool:
		return g.ValueB, nil
	case MetadataTypeString:
		return g.ValueS, nil
	case MetadataTypeRef:
		return g.ValueRef, nil
	default:
		return MetadataValue{}, errors.New("unsupported metadata type")
	}
}

// addArgOrNull adds a new argument to the list of arguments
// this helps us making it NULL if it's equal to its zero value
// The caveat is that we won't be able to pass the zero value of the type. Issues with this?
func addArgOrNull(oldArgs []any, newArg any, nullIfZero bool) []any {
	if nullIfZero && reflect.ValueOf(newArg).IsZero() {
		return append(oldArgs, nil)
	}

	return append(oldArgs, newArg)
}

func (s DeployedStream) getMetadata(ctx context.Context, params GetMetadataParams) ([]GetMetadataResult, error) {

	var args []any

	args = addArgOrNull(args, params.Key.String(), false)
	args = addArgOrNull(args, params.OnlyLatest, false)
	// just add null if ref is empty, because it's optional
	args = addArgOrNull(args, params.Ref, true)

	res, err := s._client.Call(ctx, s.DBID, "get_metadata", args)
	if err != nil {
		return nil, err
	}

	return DecodeCallResult[GetMetadataResult](res)
}

// ## Write procedures

type metadataInput struct {
	Key   MetadataKey
	Value MetadataValue
}

func (s DeployedStream) BatchInsertMetadata(ctx context.Context, inputs []metadataInput) (transactions.TxHash, error) {
	var tuples [][]any
	for _, input := range inputs {
		valType := input.Key.GetType()
		valStr, err := valType.StringFromValue(input.Value)

		if err != nil {
			return nil, err
		}

		tuples = append(tuples, []any{input.Key.String(), valStr, string(valType)})
	}

	return s._client.Execute(ctx, s.DBID, "insert_metadata", tuples)
}

func (s DeployedStream) insertMetadata(ctx context.Context, key MetadataKey, value MetadataValue) (transactions.TxHash, error) {
	return s.BatchInsertMetadata(ctx, []metadataInput{{key, value}})
}

func (s DeployedStream) disableMetadata(ctx context.Context, rowId string) (transactions.TxHash, error) {
	return s._client.Execute(ctx, s.DBID, "disable_metadata", [][]any{{rowId}})
}

func (s DeployedStream) InitializeStream(ctx context.Context) (transactions.TxHash, error) {
	return s._client.Execute(ctx, s.DBID, "init", nil)
}
