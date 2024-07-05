package tsn_api

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/client"
	kwilUtils "github.com/kwilteam/kwil-db/core/utils"
	"github.com/truflation/tsn-sdk/internal/utils"
	"strings"
)

// ## Initializations

type Stream struct {
	StreamId  utils.StreamId
	_type     StreamType
	_deployer []byte
	_owner    []byte
	DBID      string
	_client   client.Client
}

type NewStreamOptions struct {
	Client   client.Client
	StreamId utils.StreamId
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

func (s Stream) GetSchema(ctx context.Context) (*types.Schema, error) {
	return s._client.GetSchema(ctx, s.DBID)
}

func (s Stream) GetType(ctx context.Context) (StreamType, error) {
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

func (s Stream) GetStreamOwner(ctx context.Context) ([]byte, error) {
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
