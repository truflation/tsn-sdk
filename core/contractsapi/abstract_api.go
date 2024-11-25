package contractsapi

import (
	"context"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/pkg/errors"
	"github.com/trufnetwork/truf-node-sdk-go/core/types"
	"github.com/trufnetwork/truf-node-sdk-go/core/util"
)

func (s *Stream) AllowReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, types.AllowReadWalletKey, types.NewMetadataValue(wallet.Address()))
}

func (s *Stream) DisableReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error) {
	return s.disableMetadataByRef(ctx, types.AllowReadWalletKey, wallet.Address())
}

func (s *Stream) AllowComposeStream(ctx context.Context, locator types.StreamLocator) (transactions.TxHash, error) {
	streamId := locator.StreamId
	dbid := utils.GenerateDBID(streamId.String(), locator.DataProvider.Bytes())
	return s.insertMetadata(ctx, types.AllowComposeStreamKey, types.NewMetadataValue(dbid))
}

func (s *Stream) DisableComposeStream(ctx context.Context, locator types.StreamLocator) (transactions.TxHash, error) {
	dbid := utils.GenerateDBID(locator.StreamId.String(), locator.DataProvider.Bytes())
	return s.disableMetadataByRef(ctx, types.AllowComposeStreamKey, dbid)
}

func (s *Stream) GetComposeVisibility(ctx context.Context) (*util.VisibilityEnum, error) {
	results, err := s.getMetadata(ctx, getMetadataParams{
		Key:        types.ComposeVisibilityKey,
		OnlyLatest: true,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// there can be no visibility set if
	// - it's not initialized
	// - all values are disabled
	if len(results) == 0 {
		return nil, nil
	}

	value, err := results[0].GetValueByKey(types.ComposeVisibilityKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	visibility, err := util.NewVisibilityEnum(value.(int))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &visibility, nil
}

func (s *Stream) SetComposeVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, types.ComposeVisibilityKey, types.NewMetadataValue(int(visibility)))
}

func (s *Stream) GetReadVisibility(ctx context.Context) (*util.VisibilityEnum, error) {
	values, err := s.getMetadata(ctx, getMetadataParams{
		Key:        types.ReadVisibilityKey,
		OnlyLatest: true,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// there can be no visibility set if
	// - it's not initialized
	// - all values are disabled
	if len(values) == 0 {
		return nil, nil
	}

	visibility, err := util.NewVisibilityEnum(values[0].ValueI)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &visibility, nil
}

func (s *Stream) GetAllowedReadWallets(ctx context.Context) ([]util.EthereumAddress, error) {
	results, err := s.getMetadata(ctx, getMetadataParams{
		Key: types.AllowReadWalletKey,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	wallets := make([]util.EthereumAddress, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(types.AllowReadWalletKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		address, err := util.NewEthereumAddressFromString(value.(string))
		if err != nil {
			return nil, errors.WithStack(err)
		}

		wallets[i] = address
	}

	return wallets, nil
}

func (s *Stream) GetAllowedComposeStreams(ctx context.Context) ([]types.StreamLocator, error) {
	results, err := s.getMetadata(ctx, getMetadataParams{
		Key: types.AllowComposeStreamKey,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	streams := make([]types.StreamLocator, len(results))

	for i, result := range results {
		value, err := result.GetValueByKey(types.AllowComposeStreamKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// dbids are stored, not streamIds and data providers
		// so we get this, then later we query the schema
		dbid, ok := value.(string)
		if !ok {
			return nil, errors.New("invalid value type")
		}

		loc, err := s._client.GetSchema(ctx, dbid)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		streamId, err := util.NewStreamId(loc.Name)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		owner, err := util.NewEthereumAddressFromString(loc.Owner.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		streams[i] = types.StreamLocator{
			StreamId:     *streamId,
			DataProvider: owner,
		}
	}

	return streams, nil
}

func (s *Stream) SetReadVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, types.ReadVisibilityKey, types.NewMetadataValue(int(visibility)))
}

func (s *Stream) SetDefaultBaseDate(ctx context.Context, baseDate string) (transactions.TxHash, error) {
	return s.insertMetadata(ctx, types.DefaultBaseDateKey, types.NewMetadataValue(baseDate))
}

var MetadataValueNotFound = errors.New("metadata value not found")

func (s *Stream) disableMetadataByRef(ctx context.Context, key types.MetadataKey, ref string) (transactions.TxHash, error) {
	metadataList, err := s.getMetadata(ctx, getMetadataParams{
		Key:        key,
		OnlyLatest: true,
		Ref:        ref,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(metadataList) == 0 {
		return nil, MetadataValueNotFound
	}

	return s.disableMetadata(ctx, metadataList[0].RowId)
}
