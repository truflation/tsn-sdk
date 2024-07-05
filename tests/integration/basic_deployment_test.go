package integration

import (
	"context"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/internal/contractsapi"
	"github.com/truflation/tsn-sdk/internal/tsnclient"
	tsntype "github.com/truflation/tsn-sdk/internal/types"
	"github.com/truflation/tsn-sdk/internal/util"
	"testing"
	"time"
)

const TestPrivateKey = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
const TestKwilProvider = "http://localhost:8484"

func TestBasicDeployment(t *testing.T) {
	ctx := context.Background()

	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	assertNoErrorOrFail(t, err, "Failed to parse private key")

	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, err := tsnclient.NewClient(ctx, TestKwilProvider, tsnclient.WithSigner(signer))
	assertNoErrorOrFail(t, err, "Failed to create Kwil client")

	streamId := util.GenerateStreamId("test-basic-deployment")

	t.Cleanup(func() {
		destroyResult, err := tsnClient.DestroyStream(ctx, streamId)
		assert.NoError(t, err, "Failed to destroy stream")
		expectSuccessTx(t, ctx, tsnClient, destroyResult)
	})

	t.Run("Deploy Primitive, insert record and query", func(t *testing.T) {

		// Deploy a primitive stream
		deployTxHash, err := tsnClient.DeployStream(ctx, streamId, contractsapi.StreamTypePrimitive)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to deploy stream")
		expectSuccessTx(t, ctx, tsnClient, deployTxHash)

		// Load the deployed stream
		deployedStream, err := tsnClient.LoadStream(ctx, streamId)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to load stream")

		// Initialize the stream
		txHashInit, err := deployedStream.InitializeStream(ctx)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to initialize stream")
		expectSuccessTx(t, ctx, tsnClient, txHashInit)

		// Create a deployed primitive stream
		deployedPrimitiveStream, err := deployedStream.ToPrimitiveStream(ctx)
		// expect ok
		assertNoErrorOrFail(t, err, "Failed to create deployed primitive stream")

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

/*
 * ----------------------------------------------------------------------------
 * 		Helper functions
 * ----------------------------------------------------------------------------
 */

// unsafeParseDate is a helper function to parse a date string into a civil.Date, panicking on error.
func unsafeParseDate(dateStr string) *civil.Date {
	date, err := civil.ParseDate(dateStr)
	if err != nil {
		panic(err)
	}
	return &date
}

// expectSuccessTx waits for a transaction to be successful, failing the test if it fails.
func expectSuccessTx(t *testing.T, ctx context.Context, client tsntype.Client, txHash transactions.TxHash) {
	txRes, err := client.WaitForTx(ctx, txHash, time.Second)
	assertNoErrorOrFail(t, err, "Transaction failed")
	if !assert.Equal(t, transactions.CodeOk, transactions.TxCode(txRes.TxResult.Code), "Transaction code not OK: %s", txRes.TxResult.Log) {
		t.FailNow()
	}
}

// assertNoErrorOrFail asserts that an error is nil, failing the test if it is not.
func assertNoErrorOrFail(t *testing.T, err error, msg string) {
	if !assert.NoError(t, err, msg) {
		t.FailNow()
	}
}
