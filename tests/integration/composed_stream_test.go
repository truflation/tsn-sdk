package integration

import (
	"context"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/internal/tsnclient"
	"github.com/truflation/tsn-sdk/internal/types"
	"github.com/truflation/tsn-sdk/internal/util"
	"testing"
)

func TestComposedStream(t *testing.T) {
	ctx := context.Background()

	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	signerAddress, err := util.NewEthereumAddressFromBytes(signer.Identity())
	assertNoErrorOrFail(t, err, "Failed to create signer address")

	streamId := util.GenerateStreamId("test-composed-stream")

	childAStreamId := util.GenerateStreamId("test-composed-stream-child-a")
	childBStreamId := util.GenerateStreamId("test-composed-stream-child-b")

	allStreamIds := []util.StreamId{streamId, childAStreamId, childBStreamId}

	t.Cleanup(func() {
		for _, id := range allStreamIds {
			destroyResult, err := tsnClient.DestroyStream(ctx, id)
			assertNoErrorOrFail(t, err, "Failed to destroy stream")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, destroyResult)
		}
	})

	t.Run("Basic Composed Stream", func(t *testing.T) {
		// Deploy a primitive stream
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, types.StreamTypeComposed)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to deploy stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployTxHash)

		// Load the deployed stream
		deployedComposedStream, err := tsnClient.LoadComposedStream(streamId)

		// Initialize the stream
		txHashInit, err := deployedComposedStream.InitializeStream(ctx)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to initialize stream")
		waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInit)

		// deploy child streams with data
		// | date       | childA | childB |
		// |------------|--------|--------|
		// | 2020-01-01 | 1      | 3      |
		// | 2020-01-02 | 2      | 4      |

		childrenWithData := []struct {
			id   util.StreamId
			data []types.InsertRecordInput
		}{
			{
				id: childAStreamId,
				data: []types.InsertRecordInput{
					{
						Value:     1,
						DateValue: *unsafeParseDate("2020-01-01"),
					},
					{
						Value:     2,
						DateValue: *unsafeParseDate("2020-01-02"),
					},
				},
			},
			{
				id: childBStreamId,
				data: []types.InsertRecordInput{
					{
						Value:     3,
						DateValue: *unsafeParseDate("2020-01-01"),
					},
					{
						Value:     4,
						DateValue: *unsafeParseDate("2020-01-02"),
					},
				},
			},
		}

		// deploy and insert data to child streams
		for _, child := range childrenWithData {
			deployChildTxHash, err := tsnClient.DeployStream(ctx, child.id, types.StreamTypePrimitive)
			assertNoErrorOrFail(t, err, "Failed to deploy child stream")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, deployChildTxHash)

			deployedChildStream, err := tsnClient.LoadPrimitiveStream(child.id)
			assertNoErrorOrFail(t, err, "Failed to load child stream")

			txHashInitChild, err := deployedChildStream.InitializeStream(ctx)
			assertNoErrorOrFail(t, err, "Failed to initialize child stream")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInitChild)

			txHashInsert, err := deployedChildStream.InsertRecords(ctx, child.data)
			assertNoErrorOrFail(t, err, "Failed to insert record")
			waitTxToBeMinedWithSuccess(t, ctx, tsnClient, txHashInsert)
		}

		// deploy taxonomies to the composed stream
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

		// describe taxonomies
		taxonomies, err := deployedComposedStream.DescribeTaxonomies(ctx, types.DescribeTaxonomiesParams{
			LatestVersion: true,
		})
		assertNoErrorOrFail(t, err, "Failed to describe taxonomies")
		assert.Equal(t, 2, len(taxonomies))

		// query the composed stream
		records, err := deployedComposedStream.GetRecords(ctx, types.GetRecordsInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2020-01-02"),
		})

		assertNoErrorOrFail(t, err, "Failed to get records")

		assert.Equal(t, 2, len(records))

		var checkRecord = func(record types.StreamRecord, expectedValue float64) {
			val, err := record.Value.Float64()
			assertNoErrorOrFail(t, err, "Failed to parse value")
			assert.Equal(t, expectedValue, val)
		}

		// (v1 * w1 + v2 * w2 ) / (w1 + w2)
		// ( 1 *  1 +  3 *  2 ) / ( 1 +  2) = 7 / 3 = 2.333
		// ( 2 *  1 +  4 *  2 ) / ( 1 +  2) = 10 / 3 = 3.333
		checkRecord(records[0], 2.333)
		checkRecord(records[1], 3.333)
	})
}
