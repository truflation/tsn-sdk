package integration

import (
	"context"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"testing"
	"time"
)

func TestDeployComposedStreamsWithTaxonomy(t *testing.T) {
	ctx := context.Background()

	// Parse the private key for authentication
	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	// Create a signer using the parsed private key
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	// Generate unique stream IDs and locators
	primitiveStreamId := util.GenerateStreamId("test-primitive-stream-one")
	primitiveStreamId2 := util.GenerateStreamId("test-primitive-stream-two")
	composedStreamId := util.GenerateStreamId("test-composed-stream")

	// Cleanup function to destroy the streams and contracts after test completion
	t.Cleanup(func() {
		allStreamIds := []util.StreamId{primitiveStreamId, composedStreamId, primitiveStreamId2}
		for _, id := range allStreamIds {
			destroyResult, err := tsnClient.DestroyStream(ctx, id)
			assertNoErrorOrFail(t, err, "Failed to destroy stream")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, destroyResult)
		}
	})

	// Deploy a primitive stream
	deployTxHash, err := tsnClient.DeployStream(ctx, primitiveStreamId, types.StreamTypePrimitive)
	assertNoErrorOrFail(t, err, "Failed to deploy primitive stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

	// Deploy a second primitive stream
	deployTxHash, err = tsnClient.DeployStream(ctx, primitiveStreamId2, types.StreamTypePrimitive)
	assertNoErrorOrFail(t, err, "Failed to deploy primitive stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

	// Deploy a composed stream using utility function
	err = tsnClient.DeployComposedStreamWithTaxonomy(ctx, composedStreamId, types.Taxonomy{
		TaxonomyItems: []types.TaxonomyItem{
			{
				ChildStream: tsnClient.OwnStreamLocator(primitiveStreamId),
				Weight:      50,
			},
			{
				ChildStream: tsnClient.OwnStreamLocator(primitiveStreamId2),
				Weight:      50,
			},
		},
	})
	assertNoErrorOrFail(t, err, "Failed to deploy composed stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

	// List all streams
	streams, err := tsnClient.GetAllStreams(ctx, types.GetAllStreamsInput{})
	assertNoErrorOrFail(t, err, "Failed to list all streams")

	// Check that only the primitive and composed streams are listed
	expectedStreamIds := map[util.StreamId]bool{
		primitiveStreamId:  true,
		composedStreamId:   true,
		primitiveStreamId2: true,
	}

	for _, stream := range streams {
		// this will only be true if the database is clean from start
		//assert.True(t, expectedStreamIds[stream.StreamId], "Unexpected stream listed: %s", stream.StreamId)
		delete(expectedStreamIds, stream.StreamId)
	}

	// Ensure all expected streams were found
	assert.Empty(t, expectedStreamIds, "Not all expected streams were listed")

	// insert a record to primitiveStreamId and primitiveStreamId2
	// Load the primitive stream
	primitiveStream, err := tsnClient.LoadPrimitiveStream(tsnClient.OwnStreamLocator(primitiveStreamId))
	assertNoErrorOrFail(t, err, "Failed to load primitive stream")

	// Initialize the stream primitiveStreamId
	initTxHash, err := primitiveStream.InitializeStream(ctx)
	assertNoErrorOrFail(t, err, "Failed to initialize stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, initTxHash)

	// insert a record to primitiveStreamId
	insertTxHash, err := primitiveStream.InsertRecords(ctx, []types.InsertRecordInput{
		{
			DateValue: civil.DateOf(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			Value:     10,
		},
	})
	assertNoErrorOrFail(t, err, "Failed to insert record")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, insertTxHash)

	// Load the second primitive stream
	primitiveStream2, err := tsnClient.LoadPrimitiveStream(tsnClient.OwnStreamLocator(primitiveStreamId2))
	assertNoErrorOrFail(t, err, "Failed to load primitive stream")

	// Initialize the stream primitiveStreamId2
	initTxHash, err = primitiveStream2.InitializeStream(ctx)
	assertNoErrorOrFail(t, err, "Failed to initialize stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, initTxHash)

	// insert a record to primitiveStreamId2
	insertTxHash, err = primitiveStream2.InsertRecords(ctx, []types.InsertRecordInput{
		{
			DateValue: civil.DateOf(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			Value:     20,
		},
	})
	assertNoErrorOrFail(t, err, "Failed to insert record")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, insertTxHash)

	// Load the composed stream
	composedStream, err := tsnClient.LoadComposedStream(tsnClient.OwnStreamLocator(composedStreamId))
	assertNoErrorOrFail(t, err, "Failed to load composed stream")

	// Get records from the composed stream
	records, err := composedStream.GetRecord(ctx, types.GetRecordInput{})
	assertNoErrorOrFail(t, err, "Failed to get records")
	assert.Equal(t, 1, len(records), "Unexpected number of records")
	assert.Equal(t, "15.000000000000000000", records[0].Value.String(), "10 * 50/100 + 20 * 50/100 != 15")

	////
	// Negative test cases
	////

	// Deploy a composed stream with a non-existent child stream
	err = tsnClient.DeployComposedStreamWithTaxonomy(ctx, composedStreamId, types.Taxonomy{
		TaxonomyItems: []types.TaxonomyItem{
			{
				ChildStream: tsnClient.OwnStreamLocator(util.GenerateStreamId("non-existent-stream")),
				Weight:      50,
			},
		},
	})
	assert.Error(t, err, "Expected error when deploying composed stream with non-existent child stream")

	// Deploy a composed stream with already deployed stream
	err = tsnClient.DeployComposedStreamWithTaxonomy(ctx, composedStreamId, types.Taxonomy{})
	assert.Error(t, err, "Expected error when deploying already deployed stream")
}
