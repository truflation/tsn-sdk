package integration

import (
	"context"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"github.com/truflation/tsn-sdk/tests/integration/assets"
	"testing"
)

func TestListAllStreams(t *testing.T) {
	ctx := context.Background()

	// Parse the private key for authentication
	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	// Create a signer using the parsed private key
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	// Generate unique stream IDs and locators
	primitiveStreamId := util.GenerateStreamId("test-allstreams-primitive-stream")
	composedStreamId := util.GenerateStreamId("test-allstreams-composed-stream")
	notAStreamName := "not_a_stream"

	// Cleanup function to destroy the streams and contracts after test completion
	t.Cleanup(func() {
		allStreamIds := []util.StreamId{primitiveStreamId, composedStreamId}
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

	// Deploy a composed stream
	deployTxHash, err = tsnClient.DeployStream(ctx, composedStreamId, types.StreamTypeComposed)
	assertNoErrorOrFail(t, err, "Failed to deploy composed stream")
	waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

	// Deploy a non-stream contract
	notAStreamSchema, err := parse.Parse(assets.NotAStreamContent)
	notAStreamSchema.Name = notAStreamName
	assertNoErrorOrFail(t, err, "Failed to parse non-stream contract content")

	// Cleanup function to destroy the non-stream contract after test completion
	t.Cleanup(func() {
		_, err := tsnClient.GetKwilClient().DropDatabase(ctx, notAStreamName, client.WithSyncBroadcast(true))
		assertNoErrorOrFail(t, err, "Failed to destroy non-stream contract")
	})

	_, err = tsnClient.GetKwilClient().DeployDatabase(ctx, notAStreamSchema, client.WithSyncBroadcast(true))
	assertNoErrorOrFail(t, err, "Failed to deploy non-stream contract")

	// List all streams
	streams, err := tsnClient.GetAllStreams(ctx)
	assertNoErrorOrFail(t, err, "Failed to list all streams")

	// Check that only the primitive and composed streams are listed
	expectedStreamIds := map[util.StreamId]bool{
		primitiveStreamId: true,
		composedStreamId:  true,
	}

	for _, stream := range streams {
		// this will only be true if the database is clean from start
		//assert.True(t, expectedStreamIds[stream.StreamId], "Unexpected stream listed: %s", stream.StreamId)
		delete(expectedStreamIds, stream.StreamId)
	}

	// Ensure all expected streams were found
	assert.Empty(t, expectedStreamIds, "Not all expected streams were listed")
}
