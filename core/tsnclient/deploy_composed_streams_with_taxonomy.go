package tsnclient

import (
	"context"
	"fmt"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"log"
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
	log.Printf("Deployed stream %s, with txHash %s\n", streamId.String(), txHashCreate.Hex())

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
	log.Printf("Initialized stream %s, with txHash %s\n", streamId.String(), txHashInit.Hex())

	// set the taxonomy
	txHashSet, err := stream.SetTaxonomy(ctx, taxonomy)
	if err != nil {
		return err
	}

	_, err = c.WaitForTx(ctx, txHashSet, time.Second*10)
	if err != nil {
		return err
	}
	log.Printf("Set taxonomy for stream %s, with txHash %s\n", streamId.String(), txHashSet.Hex())

	return nil
}
