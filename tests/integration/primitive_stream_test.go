package integration

import (
	"context"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"testing"
)

// TestPrimitiveStream demonstrates the process of deploying, initializing, writing to,
// and reading from a primitive stream in TSN using the TSN SDK.
func TestPrimitiveStream(t *testing.T) {
	ctx := context.Background()

	// Parse the private key
	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	// Create a signer and client
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	// Generate a stream ID and locator
	streamId := util.GenerateStreamId("test-primitive-stream")
	streamLocator := tsnClient.OwnStreamLocator(streamId)

	// Cleanup function to destroy the stream after test completion
	t.Cleanup(func() {
		destroyResult, err := tsnClient.DestroyStream(ctx, streamId)
		assertNoErrorOrFail(t, err, "Failed to destroy stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, destroyResult)
	})

	// Subtest for deploying, initializing, writing to, and reading from a primitive stream
	t.Run("DeploymentWriteAndReadOperations", func(t *testing.T) {
		// Deploy a primitive stream
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, types.StreamTypePrimitive)
		assertNoErrorOrFail(t, err, "Failed to deploy stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

		// Load the deployed stream
		deployedPrimitiveStream, err := tsnClient.LoadPrimitiveStream(streamLocator)
		assertNoErrorOrFail(t, err, "Failed to load stream")

		// Initialize the stream
		txHashInit, err := deployedPrimitiveStream.InitializeStream(ctx)
		assertNoErrorOrFail(t, err, "Failed to initialize stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInit)

		// Insert a record into the stream
		txHash, err := deployedPrimitiveStream.InsertRecords(ctx, []types.InsertRecordInput{
			{
				Value:     1,
				DateValue: *unsafeParseDate("2020-01-01"),
			},
		})
		assertNoErrorOrFail(t, err, "Failed to insert record")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHash)

		// Query records from the stream
		records, err := deployedPrimitiveStream.GetRecords(ctx, types.GetRecordsInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2021-01-01"),
		})
		assertNoErrorOrFail(t, err, "Failed to query records")

		// Verify the record's content
		assert.Len(t, records, 1, "Expected exactly one record")
		assert.Equal(t, "1.000", records[0].Value.String(), "Unexpected record value")
		assert.Equal(t, "2020-01-01", records[0].DateValue.String(), "Unexpected record date")
	})
}
