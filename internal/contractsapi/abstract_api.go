package contractsapi

import (
	"context"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/pkg/errors"
	"github.com/truflation/tsn-sdk/internal/util"
)

func (s *Stream) AllowReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, AllowReadWalletKey, NewMetadataValue(wallet.Address()))
}

func (s *Stream) DisableReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error) {
	return s.disableMetadataByRef(ctx, AllowReadWalletKey, wallet.Address())
}

func (s *Stream) AllowComposeStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, AllowComposeStreamKey, NewMetadataValue(streamId.String()))
}

func (s *Stream) DisableComposeStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error) {
	return s.disableMetadataByRef(ctx, AllowComposeStreamKey, streamId.String())
}

func (s *Stream) GetComposeVisibility(ctx context.Context) (*util.VisibilityEnum, error) {
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

	visibility, err := util.NewVisibilityEnum(value.(int))

	if err != nil {
		return nil, err
	}

	return &visibility, nil
}

func (s *Stream) SetComposeVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, ComposeVisibilityKey, NewMetadataValue(int(visibility)))
}

func (s *Stream) GetReadVisibility(ctx context.Context) (*util.VisibilityEnum, error) {
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

	visibility, err := util.NewVisibilityEnum(values[0].ValueI)

	if err != nil {
		return nil, err
	}

	return &visibility, nil
}

func (s *Stream) GetAllowedReadWallets(ctx context.Context) ([]util.EthereumAddress, error) {
	results, err := s.getMetadata(ctx, GetMetadataParams{
		Key: AllowReadWalletKey,
	})

	if err != nil {
		return nil, err
	}

	wallets := make([]util.EthereumAddress, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(AllowReadWalletKey)
		if err != nil {
			return nil, err
		}

		address, err := util.NewEthereumAddress(value.(string))
		if err != nil {
			return nil, err
		}

		wallets[i] = address
	}

	return wallets, nil
}

func (s *Stream) GetAllowedComposeStreams(ctx context.Context) ([]util.StreamId, error) {
	results, err := s.getMetadata(ctx, GetMetadataParams{
		Key: AllowComposeStreamKey,
	})

	if err != nil {
		return nil, err
	}

	streams := make([]util.StreamId, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(AllowComposeStreamKey)
		if err != nil {
			return nil, err
		}

		streamId, err := util.NewStreamId(value.(string))

		if err != nil {
			return nil, err
		}

		streams[i] = *streamId
	}

	return streams, nil
}

func (s *Stream) SetReadVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, ReadVisibilityKey, NewMetadataValue(int(visibility)))
}

var MetadataValueNotFound = errors.New("metadata value not found")

func (s *Stream) disableMetadataByRef(ctx context.Context, key MetadataKey, ref string) (transactions.TxHash, error) {
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