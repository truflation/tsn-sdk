package types

import (
	"context"
	kwilClientPkg "github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/truflation/tsn-sdk/core/util"
	"time"
)

type Client interface {
	// WaitForTx waits for the transaction to be mined by TSN
	WaitForTx(ctx context.Context, txHash transactions.TxHash, interval time.Duration) (*transactions.TcTxQueryResponse, error)
	// GetKwilClient returns the kwil client used by the client
	GetKwilClient() *kwilClientPkg.Client
	// DeployStream deploys a new stream
	DeployStream(ctx context.Context, streamId util.StreamId, streamType StreamType) (transactions.TxHash, error)
	// DestroyStream destroys a stream
	DestroyStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error)
	// LoadStream loads a already deployed stream, permitting its API usage
	LoadStream(stream StreamLocator) (IStream, error)
	// LoadPrimitiveStream loads a already deployed primitive stream, permitting its API usage
	LoadPrimitiveStream(stream StreamLocator) (IPrimitiveStream, error)
	// LoadComposedStream loads a already deployed composed stream, permitting its API usage
	LoadComposedStream(stream StreamLocator) (IComposedStream, error)
	/*
	 * utils for the client
	 */
	// Create a new stream locator
	OwnStreamLocator(streamId util.StreamId) StreamLocator
	// Address of the signer used by the client
	Address() util.EthereumAddress
	// GetAllStreams returns all streams from the TSN network
	GetAllStreams(ctx context.Context, input GetAllStreamsInput) ([]StreamLocator, error)
	// DeployComposedStreamWithTaxonomy deploys a composed stream with a taxonomy
	DeployComposedStreamWithTaxonomy(ctx context.Context, streamId util.StreamId, taxonomy Taxonomy) error
}

type GetAllStreamsInput struct {
	Owner []byte
}
