package tsn_api

import (
	"context"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/pkg/errors"
	"github.com/truflation/tsn-sdk/internal/utils"
)

func (s DeployedStream) AllowReadWallet(ctx context.Context, wallet utils.EthereumAddress) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, AllowReadWalletKey, NewMetadataValue(wallet.Address()))
}

func (s DeployedStream) DisableReadWallet(ctx context.Context, wallet utils.EthereumAddress) (transactions.TxHash, error) {
	return s.disableMetadataByRef(ctx, AllowReadWalletKey, wallet.Address())
}

func (s DeployedStream) AllowComposeStream(ctx context.Context, streamId utils.StreamId) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, AllowComposeStreamKey, NewMetadataValue(streamId.String()))
}

func (s DeployedStream) DisableComposeStream(ctx context.Context, streamId utils.StreamId) (transactions.TxHash, error) {
	return s.disableMetadataByRef(ctx, AllowComposeStreamKey, streamId.String())
}

func (s DeployedStream) GetComposeVisibility(ctx context.Context) (*utils.VisibilityEnum, error) {
	results, err := s.getMetadata(ctx, GetMetadataParams{
		Key:        ComposeVisibilityKey,
		OnlyLatest: true,
	})

	if err != nil {
		return nil, err
	}

	// there can be no visibility set if
	// - it's not initialized
	// - all values are disabled
	if len(results) == 0 {
		return nil, nil
	}

	value, err := results[0].GetValueByKey(ComposeVisibilityKey)
	if err != nil {
		return nil, err
	}

	visibility, err := utils.NewVisibilityEnum(value.(int))

	if err != nil {
		return nil, err
	}

	return &visibility, nil
}

func (s DeployedStream) SetComposeVisibility(ctx context.Context, visibility utils.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, ComposeVisibilityKey, NewMetadataValue(int(visibility)))
}

func (s DeployedStream) GetReadVisibility(ctx context.Context) (*utils.VisibilityEnum, error) {
	values, err := s.getMetadata(ctx, GetMetadataParams{
		Key:        ReadVisibilityKey,
		OnlyLatest: true,
	})

	if err != nil {
		return nil, err
	}

	// there can be no visibility set if
	// - it's not initialized
	// - all values are disabled
	if len(values) == 0 {
		return nil, nil
	}

	visibility, err := utils.NewVisibilityEnum(values[0].ValueI)

	if err != nil {
		return nil, err
	}

	return &visibility, nil
}

func (s DeployedStream) GetAllowedReadWallets(ctx context.Context) ([]utils.EthereumAddress, error) {
	results, err := s.getMetadata(ctx, GetMetadataParams{
		Key: AllowReadWalletKey,
	})

	if err != nil {
		return nil, err
	}

	wallets := make([]utils.EthereumAddress, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(AllowReadWalletKey)
		if err != nil {
			return nil, err
		}

		address, err := utils.NewEthereumAddress(value.(string))
		if err != nil {
			return nil, err
		}

		wallets[i] = address
	}

	return wallets, nil
}

func (s DeployedStream) GetAllowedComposeStreams(ctx context.Context) ([]utils.StreamId, error) {
	results, err := s.getMetadata(ctx, GetMetadataParams{
		Key: AllowComposeStreamKey,
	})

	if err != nil {
		return nil, err
	}

	streams := make([]utils.StreamId, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(AllowComposeStreamKey)
		if err != nil {
			return nil, err
		}

		streamId, err := utils.NewStreamId(value.(string))

		if err != nil {
			return nil, err
		}

		streams[i] = *streamId
	}

	return streams, nil
}

func (s DeployedStream) SetReadVisibility(ctx context.Context, visibility utils.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, ReadVisibilityKey, NewMetadataValue(int(visibility)))
}

var MetadataValueNotFound = errors.New("metadata value not found")

func (s DeployedStream) disableMetadataByRef(ctx context.Context, key MetadataKey, ref string) (transactions.TxHash, error) {
	metadataList, err := s.getMetadata(ctx, GetMetadataParams{
		Key:        key,
		OnlyLatest: true,
		Ref:        ref,
	})

	if err != nil {
		return nil, err
	}

	if len(metadataList) == 0 {
		return nil, MetadataValueNotFound
	}

	return s.disableMetadata(ctx, metadataList[0].RowId)
}
