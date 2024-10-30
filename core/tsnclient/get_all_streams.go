package tsnclient

import (
	"context"
	kwiltypes "github.com/kwilteam/kwil-db/core/types"
	tsntypes "github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"log"
)

// GetAllStreams returns all streams from the TSN network
func (c *Client) GetAllStreams(ctx context.Context, input tsntypes.GetAllStreamsInput) ([]tsntypes.StreamLocator, error) {
	kwilClient := c.GetKwilClient()

	// get all deployed contracts
	contracts, err := kwilClient.ListDatabases(ctx, input.Owner)

	if err != nil {
		return nil, err
	}

	// create a list of stream locators
	streamLocators := make([]tsntypes.StreamLocator, 0)

	// iterate over all contracts
	// if the contract is a stream, add it to the list
	for _, contract := range contracts {
		schema, err := kwilClient.GetSchema(ctx, contract.DBID)
		if err != nil {
			return nil, err
		}

		isStream, err := getIsStream(schema)

		if isStream {
			streamId, err := util.NewStreamId(contract.Name)
			if err != nil {
				// in this case, contracts such as system contract won't have a valid streamId as name, we just continue it
				continue
			}
			dataProvider, err := util.NewEthereumAddressFromBytes(contract.Owner)
			if err != nil {
				// we should return the error in this case. Every owner should be an ethereum address
				return nil, err
			}

			streamLocators = append(streamLocators, tsntypes.StreamLocator{
				StreamId:     *streamId,
				DataProvider: dataProvider,
			})
		}
	}

	return streamLocators, nil
}

func (c *Client) GetAllInitializedStreams(ctx context.Context, input tsntypes.GetAllStreamsInput) ([]tsntypes.StreamLocator, error) {
	kwilClient := c.GetKwilClient()

	// get all deployed contracts
	contracts, err := kwilClient.ListDatabases(ctx, input.Owner)

	if err != nil {
		return nil, err
	}

	// create a list of stream locators
	streamLocators := make([]tsntypes.StreamLocator, 0)

	// iterate over all contracts
	// if the contract is a stream, add it to the list
	for _, contract := range contracts {
		schema, err := kwilClient.GetSchema(ctx, contract.DBID)
		if err != nil {
			return nil, err
		}

		isStream, err := getIsStream(schema)

		if isStream {
			streamId, err := util.NewStreamId(contract.Name)
			if err != nil {
				// in this case, contracts such as system contract won't have a valid streamId as name, we just continue it
				continue
			}
			dataProvider, err := util.NewEthereumAddressFromBytes(contract.Owner)
			if err != nil {
				// we should return the error in this case. Every owner should be an ethereum address
				return nil, err
			}

			streamLocator := tsntypes.StreamLocator{
				StreamId:     *streamId,
				DataProvider: dataProvider,
			}

			// check if the stream is initialized by trying to load it and get its type
			deployedStream, err := c.LoadStream(streamLocator)
			if err != nil {
				// in case of error, we just continue to the next stream
				log.Printf("skipping stream %s due to error on load: %s", streamId.String(), err.Error())
				continue
			}

			// get the type of the stream
			values, err := deployedStream.GetType(ctx)
			if err != nil {
				// in case of error, we just continue to the next stream, it means the stream is not initialized
				log.Printf("skipping stream %s due to error on get type: %s", streamId.String(), err.Error())
				continue
			}

			if len(values) == 0 {
				// type can't ever be disabled
				log.Printf("no type found on stream %s, check if the stream is initialized, skipping", streamId.String())
				continue
			}

			streamLocators = append(streamLocators, streamLocator)
		}
	}

	return streamLocators, nil
}

func getIsStream(schema *kwiltypes.Schema) (bool, error) {
	// we must try to differentiate streams from all other contracts. Let's improve it with time.
	// In the future there should be a clear interface that defines a stream

	// must have procedures:
	// - get_index
	// - get_record
	// - get_metadata

	procedures := schema.Procedures

	availableProcedures := make(map[string]bool)
	for _, procedure := range procedures {
		availableProcedures[procedure.Name] = true
	}

	requiredProcedures := []string{"get_index", "get_record", "get_metadata"}

	for _, requiredProcedure := range requiredProcedures {
		if !availableProcedures[requiredProcedure] {
			return false, nil
		}
	}

	return true, nil
}
