package tsnclient

import (
	"context"
	"fmt"
	"github.com/truflation/tsn-sdk/core/logging"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"go.uber.org/zap"
	"time"
)

// DeployComposedStreamsWithTaxonomy deploys a composed stream with taxonomy
func (c *Client) DeployComposedStreamWithTaxonomy(ctx context.Context, streamId util.StreamId, taxonomy types.Taxonomy) error {
	// check if the stream on taxonomies is already deployed
	for _, item := range taxonomy.TaxonomyItems {
		_, err := c.LoadStream(item.ChildStream)
		if err != nil {
			return err
		}
	}

	// check if the stream is already deployed
	_, err := c.LoadStream(c.OwnStreamLocator(streamId))
	if err == nil {
		return fmt.Errorf("stream already deployed")
	}

	// create the stream
	txHashCreate, err := c.DeployStream(ctx, streamId, types.StreamTypeComposed)
	if err != nil {
		return err
	}

	_, err = c.WaitForTx(ctx, txHashCreate, time.Second*10)
	if err != nil {
		return err
	}
	logging.Logger.Info("Deployed stream, with txHash", zap.String("streamId", streamId.String()), zap.String("txHash", txHashCreate.Hex()))

	// load the stream
	streamLocator := c.OwnStreamLocator(streamId)
	stream, err := c.LoadComposedStream(streamLocator)
	if err != nil {
		return err
	}

	// initialize the stream
	txHashInit, err := stream.InitializeStream(ctx)
	if err != nil {
		return err
	}

	_, err = c.WaitForTx(ctx, txHashInit, time.Second*10)
	if err != nil {
		return err
	}
	logging.Logger.Info("Initialized stream", zap.String("streamId", streamId.String()), zap.String("txHash", txHashInit.Hex()))

	// set the taxonomy
	txHashSet, err := stream.SetTaxonomy(ctx, taxonomy)
	if err != nil {
		return err
	}

	_, err = c.WaitForTx(ctx, txHashSet, time.Second*10)
	if err != nil {
		return err
	}
	logging.Logger.Info("Set taxonomy for stream", zap.String("streamId", streamId.String()), zap.String("txHash", txHashSet.Hex()))

	return nil
}
