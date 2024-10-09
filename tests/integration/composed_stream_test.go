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
	"time"
)

// This file contains integration tests for composed streams in the Truflation Stream Network (TSN).
// It demonstrates the process of deploying, initializing, and querying a composed stream
// that aggregates data from multiple primitive streams.

// TestComposedStream demonstrates the process of deploying, initializing, and querying
// a composed stream that aggregates data from multiple primitive streams in the TSN using the TSN SDK.
func TestComposedStream(t *testing.T) {
	ctx := context.Background()

	// Parse the private key for authentication
	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	// Create a signer using the parsed private key
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	signerAddress, err := util.NewEthereumAddressFromBytes(signer.Identity())
	assertNoErrorOrFail(t, err, "Failed to create signer address")

	// Generate a unique stream ID and locator for the composed stream and its child streams
	streamId := util.GenerateStreamId("test-composed-stream")
	streamLocator := tsnClient.OwnStreamLocator(streamId)

	childAStreamId := util.GenerateStreamId("test-composed-stream-child-a")
	childBStreamId := util.GenerateStreamId("test-composed-stream-child-b")

	allStreamIds := []util.StreamId{streamId, childAStreamId, childBStreamId}

	// Cleanup function to destroy the streams after test completion
	t.Cleanup(func() {
		for _, id := range allStreamIds {
			destroyResult, err := tsnClient.DestroyStream(ctx, id)
			assertNoErrorOrFail(t, err, "Failed to destroy stream")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, destroyResult)
		}
	})

	// Subtest for deploying, initializing, and querying the composed stream
	t.Run("DeploymentAndReadOperations", func(t *testing.T) {
		// Step 1: Deploy the composed stream
		// This creates the composed stream contract on the TSN
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, types.StreamTypeComposed)
		assertNoErrorOrFail(t, err, "Failed to deploy composed stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

		// Load the deployed composed stream
		deployedComposedStream, err := tsnClient.LoadComposedStream(streamLocator)
		assertNoErrorOrFail(t, err, "Failed to load composed stream")

		// Step 2: Initialize the composed stream
		// Initialization prepares the composed stream for data operations
		txHashInit, err := deployedComposedStream.InitializeStream(ctx)
		assertNoErrorOrFail(t, err, "Failed to initialize composed stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInit)

		// Step 3: Deploy child streams with initial data
		// Deploy two primitive child streams with initial data
		// | date       | childA | childB |
		// |------------|--------|--------|
		// | 2020-01-01 | 1      | 3      |
		// | 2020-01-02 | 2      | 4      |

		deployTestPrimitiveStreamWithData(t, ctx, tsnClient, childAStreamId, []types.InsertRecordInput{
			{Value: 1, DateValue: *unsafeParseDate("2020-01-01")},
			{Value: 2, DateValue: *unsafeParseDate("2020-01-02")},
		})

		deployTestPrimitiveStreamWithData(t, ctx, tsnClient, childBStreamId, []types.InsertRecordInput{
			{Value: 3, DateValue: *unsafeParseDate("2020-01-01")},
			{Value: 4, DateValue: *unsafeParseDate("2020-01-02")},
		})

		// Step 4: Set taxonomies for the composed stream
		// Taxonomies define the structure of the composed stream
		txHashTaxonomies, err := deployedComposedStream.SetTaxonomy(ctx, []types.TaxonomyItem{
			{
				ChildStream: types.StreamLocator{
					StreamId:     childAStreamId,
					DataProvider: signerAddress,
				},
				Weight:    1,
				StartDate: time.Date(2020, 1, 30, 0, 0, 0, 0, time.UTC).Format(time.DateOnly),
			},
			{
				ChildStream: types.StreamLocator{
					StreamId:     childBStreamId,
					DataProvider: signerAddress,
				},
				Weight: 2,
			},
		})
		assertNoErrorOrFail(t, err, "Failed to set taxonomies")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashTaxonomies)

		// Describe the taxonomies of the composed stream
		taxonomies, err := deployedComposedStream.DescribeTaxonomies(ctx, types.DescribeTaxonomiesParams{
			LatestVersion: true,
		})
		assertNoErrorOrFail(t, err, "Failed to describe taxonomies")
		assert.Equal(t, 2, len(taxonomies))
		assert.Equal(t, time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), taxonomies[0].StartDate) // should be default value because it was not set
		assert.Equal(t, time.Date(2020, 1, 30, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), taxonomies[1].StartDate)

		// Step 5: Query the composed stream for records
		// Query records within a specific date range
		records, err := deployedComposedStream.GetRecord(ctx, types.GetRecordInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2020-01-02"),
		})

		assertNoErrorOrFail(t, err, "Failed to get records")

		assert.Equal(t, 2, len(records))

		// Function to check the record values
		var checkRecord = func(record types.StreamRecord, expectedValue float64) {
			val, err := record.Value.Float64()
			assertNoErrorOrFail(t, err, "Failed to parse value")
			assert.Equal(t, expectedValue, val)
		}

		// Verify the record values
		// (v1 * w1 + v2 * w2 ) / (w1 + w2)
		// ( 1 *  1 +  3 *  2 ) / ( 1 +  2) = 7 / 3 = 2.333
		// ( 2 *  1 +  4 *  2 ) / ( 1 +  2) = 10 / 3 = 3.333
		checkRecord(records[0], 2.3333333333333335)
		checkRecord(records[1], 3.3333333333333335)

		// Step 6: Query the composed stream for index
		// Query the index within a specific date range
		index, err := deployedComposedStream.GetIndex(ctx, types.GetIndexInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2020-01-02"),
		})

		assertNoErrorOrFail(t, err, "Failed to get index")
		assert.Equal(t, 2, len(index))
		checkRecord(index[0], 100)
		checkRecord(index[1], 155.55555555555554)

		// Step 7: Query the first record from the composed stream
		// Query the first record from the composed stream
		firstRecord, err := deployedComposedStream.GetFirstRecord(ctx, types.GetFirstRecordInput{})
		assertNoErrorOrFail(t, err, "Failed to get first record")
		checkRecord(*firstRecord, 2.3333333333333335)
		assert.Equal(t, "2020-01-01", firstRecord.DateValue.String())
	})
}
