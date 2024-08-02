package tsnclient

import (
	"context"
	"github.com/go-playground/validator/v10"
	kwilClientPkg "github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	kwilClientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	tsn_api "github.com/truflation/tsn-sdk/core/contractsapi"
	clientType "github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"time"
)

type Client struct {
	Signer      auth.Signer `validate:"required"`
	logger      *log.Logger
	kwilClient  *kwilClientPkg.Client `validate:"required"`
	kwilOptions *kwilClientType.Options
}

var _ clientType.Client = (*Client)(nil)

type Option func(*Client)

func NewClient(ctx context.Context, provider string, options ...Option) (*Client, error) {
	c := &Client{}
	c.kwilOptions = kwilClientType.DefaultOptions()
	kwilClient, err := kwilClientPkg.NewClient(ctx, provider, c.kwilOptions)
	if err != nil {
		return nil, err
	}
	c.kwilClient = kwilClient
	c.Signer = kwilClient.Signer
	for _, option := range options {
		option(c)
	}

	// Validate the client
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func WithSigner(signer auth.Signer) Option {
	return func(c *Client) {
		c.kwilClient.Signer = signer
	}
}

func WithLogger(logger log.Logger) Option {
	return func(c *Client) {
		c.logger = &logger
		c.kwilOptions.Logger = logger
	}
}

func (c *Client) GetSigner() auth.Signer {
	return c.kwilClient.Signer
}

func (c *Client) WaitForTx(ctx context.Context, txHash transactions.TxHash, interval time.Duration) (*transactions.TcTxQueryResponse, error) {
	return c.kwilClient.WaitTx(ctx, txHash, interval)
}

func (c *Client) GetKwilClient() *kwilClientPkg.Client {
	return c.kwilClient
}

func (c *Client) DeployStream(ctx context.Context, streamId util.StreamId, streamType clientType.StreamType) (transactions.TxHash, error) {
	out, err := tsn_api.DeployStream(ctx, tsn_api.DeployStreamInput{
		StreamId:   streamId,
		StreamType: streamType,
		KwilClient: c.kwilClient,
		Deployer:   c.kwilClient.Signer.Identity(),
	})
	if err != nil {
		return transactions.TxHash{}, err
	}

	return out.TxHash, nil
}

func (c *Client) DestroyStream(ctx context.Context, streamId util.StreamId) (transactions.TxHash, error) {
	out, err := tsn_api.DestroyStream(ctx, tsn_api.DestroyStreamInput{
		StreamId:   streamId,
		KwilClient: c.kwilClient,
	})
	if err != nil {
		return transactions.TxHash{}, err
	}

	return out.TxHash, nil
}

func (c *Client) LoadStream(streamLocator clientType.StreamLocator) (clientType.IStream, error) {
	return tsn_api.NewStream(tsn_api.NewStreamOptions{
		Client:   c.kwilClient,
		StreamId: streamLocator.StreamId,
		Deployer: streamLocator.DataProvider.Bytes(),
	})
}

func (c *Client) LoadPrimitiveStream(streamLocator clientType.StreamLocator) (clientType.IPrimitiveStream, error) {
	return tsn_api.NewPrimitiveStream(tsn_api.NewStreamOptions{
		Client:   c.kwilClient,
		StreamId: streamLocator.StreamId,
		Deployer: streamLocator.DataProvider.Bytes(),
	})
}

func (c *Client) LoadComposedStream(streamLocator clientType.StreamLocator) (clientType.IComposedStream, error) {
	return tsn_api.NewComposedStream(tsn_api.NewStreamOptions{
		Client:   c.kwilClient,
		StreamId: streamLocator.StreamId,
		Deployer: streamLocator.DataProvider.Bytes(),
	})
}

func (c *Client) OwnStreamLocator(streamId util.StreamId) clientType.StreamLocator {
	return clientType.StreamLocator{
		StreamId:     streamId,
		DataProvider: c.Address(),
	}
}

func (c *Client) Address() util.EthereumAddress {
	address, err := util.NewEthereumAddressFromBytes(c.kwilClient.Signer.Identity())
	if err != nil {
		// should never happen
		panic(err)
	}
	return address
}
