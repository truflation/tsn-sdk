package contractsapi

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	kwilUtils "github.com/kwilteam/kwil-db/core/utils"
	"github.com/truflation/tsn-sdk/internal/util"
	"strings"
)

// ## Initializations

type Stream struct {
	StreamId     util.StreamId
	_type        StreamType
	_deployer    []byte
	_owner       []byte
	DBID         string
	_client      client.Client
	_initialized bool
	_deployed    bool
}

type NewStreamOptions struct {
	Client   client.Client
	StreamId util.StreamId
	Deployer []byte
}

const (
	ErrorStreamNotFound = "stream not found"
)

func NewStream(options NewStreamOptions) (*Stream, error) {
	optClient := options.Client
	streamId := options.StreamId
	deployer := options.Deployer

	var err error
	// if there's no deployer, let's throw an error
	if len(deployer) == 0 {
		return nil, fmt.Errorf("contract owner is required")
	}

	dbid := kwilUtils.GenerateDBID(streamId.String(), deployer)
	if err != nil {
		// if err contains "dataset not found", it means the stream is not deployed, then we return our error
		if strings.Contains(err.Error(), "dataset not found") {
			return nil, fmt.Errorf(ErrorStreamNotFound)
		}

		return nil, err
	}

	return &Stream{
		StreamId:  streamId,
		_deployer: deployer,
		DBID:      dbid,
		_client:   optClient,
	}, nil
}

func (s *Stream) ToComposedStream() (*ComposedStream, error) {
	return ComposedStreamFromStream(*s)
}

func (s *Stream) ToPrimitiveStream() (*PrimitiveStream, error) {
	return PrimitiveStreamFromStream(*s)
}

func (s *Stream) GetSchema(ctx context.Context) (*types.Schema, error) {
	return s._client.GetSchema(ctx, s.DBID)
}

func (s *Stream) GetType(ctx context.Context) (StreamType, error) {
	if s._type != "" {
		return s._type, nil
	}

	values, err := s.getMetadata(ctx, GetMetadataParams{
		Key:        "type",
		OnlyLatest: true,
	})

	if err != nil {
		return "", err
	}

	if len(values) == 0 {
		// type can't ever be disabled
		return "", fmt.Errorf("no type found, check if the stream is initialized")
	}

	switch values[0].ValueS {
	case "composed":
		s._type = StreamTypeComposed
	case "primitive":
		s._type = StreamTypePrimitive
	default:
		return "", fmt.Errorf("unknown stream type: %s", values[0].ValueS)
	}

	return s._type, nil
}

func (s *Stream) GetStreamOwner(ctx context.Context) ([]byte, error) {
	if s._owner != nil {
		return s._owner, nil
	}

	values, err := s.getMetadata(ctx, GetMetadataParams{
		Key:        "stream_owner",
		OnlyLatest: true,
	})

	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		// owner can't ever be disabled
		return nil, fmt.Errorf("no owner found (is the stream initialized?)")
	}

	s._owner, err = hex.DecodeString(values[0].ValueRef)

	if err != nil {
		return nil, err
	}

	return s._owner, nil
}

func (s *Stream) checkInitialized(ctx context.Context) error {
	if s._initialized {
		return nil
	}

	// check if is deployed
	err := s.checkDeployed(ctx)

	if err != nil {
		return err
	}

	// check if is initialized by trying to get its type
	_, err = s.GetType(ctx)
	if err != nil {
		return fmt.Errorf("check if the stream is initialized: %w", err)
	}

	s._initialized = true

	return nil
}

func (s *Stream) checkDeployed(ctx context.Context) error {
	if s._deployed {
		return nil
	}

	_, err := s.GetSchema(ctx)
	if err != nil {
		return fmt.Errorf("check if the stream is deployed: %w", err)
	}

	s._deployed = true

	return nil
}

func (s *Stream) call(ctx context.Context, method string, args []any) (*client.CallResult, error) {
	err := s.checkInitialized(ctx)
	if err != nil {
		return nil, err
	}

	return s._client.Call(ctx, s.DBID, method, args)
}

func (s *Stream) execute(ctx context.Context, method string, args [][]any) (transactions.TxHash, error) {
	err := s.checkInitialized(ctx)
	if err != nil {
		return transactions.TxHash{}, err
	}

	return s._client.Execute(ctx, s.DBID, method, args)
}