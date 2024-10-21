package contractsapi

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/core/util"
	validator "gopkg.in/validator.v2"
)

type DestroyStreamInput struct {
	StreamId   util.StreamId `validate:"nonnil"`
	KwilClient client.Client `validate:"nonnil"`
}

func (i DestroyStreamInput) Validate() error {
	return validator.Validate(i)
}

type DestroyStreamOutput struct {
	TxHash transactions.TxHash
}

// DestroyStream destroys a stream from TSN
func DestroyStream(ctx context.Context, input DestroyStreamInput) (*DestroyStreamOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	txHash, err := input.KwilClient.DropDatabase(ctx, input.StreamId.String())
	if err != nil {
		return nil, err
	}

	return &DestroyStreamOutput{
		TxHash: txHash,
	}, nil
}
