package integration

import (
	"context"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/internal/contractsapi"
	"github.com/truflation/tsn-sdk/internal/tsnclient"
	"github.com/truflation/tsn-sdk/internal/util"
	"testing"
)

func TestPrimitiveStream(t *testing.T) {
	ctx := context.Background()

	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create client")

	streamId := util.GenerateStreamId("test-primitive-stream")

	t.Cleanup(func() {
		destroyResult, err := tsnClient.DestroyStream(ctx, streamId)
		assertNoErrorOrFail(t, err, "Failed to destroy stream")
		expectSuccessTx(t, ctx, tsnClient, destroyResult)
	})

	t.Run("Basic Primitive Stream", func(t *testing.T) {
		// Deploy a primitive stream
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, contractsapi.StreamTypePrimitive)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to deploy stream")
		expectSuccessTx(t, ctx, tsnClient, deployTxHash)

		// Load the deployed stream
		deployedPrimitiveStream, err := tsnClient.LoadPrimitiveStream(streamId)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to load stream")

		// Initialize the stream
		txHashInit, err := deployedPrimitiveStream.InitializeStream(ctx)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to initialize stream")
		expectSuccessTx(t, ctx, tsnClient, txHashInit)

		txHash, err := deployedPrimitiveStream.InsertRecords(ctx, []contractsapi.InsertRecordInput{
			{
				Value:     1,
				DateValue: *unsafeParseDate("2020-01-01"),
			},
		})
		assertNoErrorOrFail(t, err, "Failed to insert record")
		expectSuccessTx(t, ctx, tsnClient, txHash)

		records, err := deployedPrimitiveStream.GetRecords(ctx, contractsapi.GetRecordsInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2021-01-01"),
		})
		assertNoErrorOrFail(t, err, "Failed to query records")

		assert.Len(t, records, 1, "Expected exactly one record")
		assert.Equal(t, "1.000", records[0].Value.String(), "Unexpected record value")
		assert.Equal(t, "2020-01-01", records[0].DateValue.String(), "Unexpected record date")
	})
}
