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

// TestComposedStream demonstrates the process of deploying, initializing, and querying
// a composed stream that aggregates data from multiple primitive streams in the TSN using the TSN SDK.
func TestComposedStream(t *testing.T) {
	ctx := context.Background()

	// Parse the private key
	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	// Create a signer and client
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	signerAddress, err := util.NewEthereumAddressFromBytes(signer.Identity())
	assertNoErrorOrFail(t, err, "Failed to create signer address")

	// Generate stream IDs for the composed stream and its child streams
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
		// Deploy the composed stream
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, types.StreamTypeComposed)
		assertNoErrorOrFail(t, err, "Failed to deploy stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

		// Load the deployed composed stream
		deployedComposedStream, err := tsnClient.LoadComposedStream(streamLocator)

		// Initialize the composed stream
		txHashInit, err := deployedComposedStream.InitializeStream(ctx)
		assertNoErrorOrFail(t, err, "Failed to initialize stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInit)

		// Deploy child streams with data
		// | date       | childA | childB |
		// |------------|--------|--------|
		// | 2020-01-01 | 1      | 3      |
		// | 2020-01-02 | 2      | 4      |

		deployTestPrimitiveStreamWithData(t, ctx, tsnClient, childAStreamId, []types.InsertRecordInput{
			{
				Value:     1,
				DateValue: *unsafeParseDate("2020-01-01"),
			},
			{
				Value:     2,
				DateValue: *unsafeParseDate("2020-01-02"),
			},
		})

		deployTestPrimitiveStreamWithData(t, ctx, tsnClient, childBStreamId, []types.InsertRecordInput{
			{
				Value:     3,
				DateValue: *unsafeParseDate("2020-01-01"),
			},
			{
				Value:     4,
				DateValue: *unsafeParseDate("2020-01-02"),
			},
		})

		// Set taxonomies for the composed stream
		// | childA | childB |
		// |--------|--------|
		// | 1      | 2      |
		txHashTaxonomies, err := deployedComposedStream.SetTaxonomy(ctx, []types.TaxonomyItem{
			{
				ChildStream: types.StreamLocator{
					StreamId:     childAStreamId,
					DataProvider: signerAddress,
				},
				Weight: 1,
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

		// Query the composed stream for records
		records, err := deployedComposedStream.GetRecords(ctx, types.GetRecordsInput{
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
		checkRecord(records[0], 2.333)
		checkRecord(records[1], 3.333)
	})
}
