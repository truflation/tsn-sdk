package contractsapi

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/pkg/errors"
	"github.com/truflation/tsn-sdk/core/contracts"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
)

type DeployStreamInput struct {
	StreamId   util.StreamId    `validate:"required"`
	StreamType types.StreamType `validate:"required"`
	KwilClient client.Client    `validate:"required"`
	Deployer   []byte           `validate:"required"`
}

type DeployStreamOutput struct {
	DeployedStream Stream
	TxHash         transactions.TxHash
}

// DeployStream deploys a stream to TSN
func DeployStream(ctx context.Context, input DeployStreamInput) (*DeployStreamOutput, error) {
	contractContent, err := GetContractContent(input)
	schema, err := parse.Parse(contractContent)
	if err != nil {
		return nil, err
	}

	schema.Name = input.StreamId.String()

	txHash, err := input.KwilClient.DeployDatabase(ctx, schema)
	if err != nil {
		return nil, err
	}

	options := NewStreamOptions{
		Client:   input.KwilClient,
		StreamId: input.StreamId,
		Deployer: input.Deployer,
	}

	deployedStream, err := NewStream(options)
	if err != nil {
		return nil, err
	}

	return &DeployStreamOutput{
		DeployedStream: *deployedStream,
		TxHash:         txHash,
	}, nil
}

// GetContractContent returns the contract content based on the stream type
func GetContractContent(input DeployStreamInput) ([]byte, error) {
	switch input.StreamType {
	case types.StreamTypeComposed:
		return contracts.ComposedContractContent, nil
	case types.StreamTypePrimitive:
		return contracts.PrivateContractContent, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown stream type: %v", input.StreamType))
	}
}