package integration

import (
	"context"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/assert"
	tsn_api "github.com/truflation/tsn-sdk/internal/contracts-api"
	"github.com/truflation/tsn-sdk/internal/utils"
	"testing"
	"time"
)

const TestPrivateKey = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
const TestKwilProvider = "http://localhost:8484"

func TestBasicDeployment(t *testing.T) {
	ctx := context.Background()

	pk, err := crypto.Secp256k1PrivateKeyFromHex(TestPrivateKey)
	if !assert.NoError(t, err, "Failed to parse private key") {
		t.FailNow()
	}

	signer := &auth.EthPersonalSigner{Key: *pk}
	kwilClient, err := client.NewClient(ctx, TestKwilProvider, &clientType.Options{
		Signer: signer,
	})
	if !assert.NoError(t, err, "Failed to create Kwil client") {
		t.FailNow()
	}

	streamId := utils.GenerateStreamId("test-basic-deployment")

	t.Cleanup(func() {
		destroyResult, err := tsn_api.DestroyStream(ctx, tsn_api.DestroyStreamInput{
			StreamId:   streamId,
			KwilClient: kwilClient,
		})
		assert.NoError(t, err, "Failed to destroy stream")
		expectSuccessTx(t, ctx, *kwilClient, destroyResult.TxHash)
	})

	t.Run("Deploy Primitive, insert record and query", func(t *testing.T) {

		deployOutput, err := tsn_api.DeployStream(ctx, tsn_api.DeployStreamInput{
			StreamId:   streamId,
			StreamType: tsn_api.StreamTypePrimitive,
			KwilClient: kwilClient,
			Deployer:   signer.Identity(),
		})
		if !assert.NoError(t, err, "Failed to deploy stream") {
			t.FailNow()
		}

		expectSuccessTx(t, ctx, *kwilClient, deployOutput.TxHash)

		txHashInit, err := deployOutput.DeployedStream.InitializeStream(ctx)
		if !assert.NoError(t, err, "Failed to initialize stream") {
			t.FailNow()
		}

		expectSuccessTx(t, ctx, *kwilClient, txHashInit)

		deployedPrimitiveStream, err := tsn_api.DeployedPrimitiveStreamFromDeployedStream(ctx, deployOutput.DeployedStream)

		if !assert.NoError(t, err, "Failed to create deployed primitive stream") {
			t.FailNow()
		}

		txHash, err := deployedPrimitiveStream.InsertRecords(ctx, []tsn_api.InsertRecordInput{
			{
				Value:     1,
				DateValue: *unsafeParseDate("2020-01-01"),
			},
		})
		if !assert.NoError(t, err, "Failed to insert record") {
			t.FailNow()
		}

		expectSuccessTx(t, ctx, *kwilClient, txHash)

		records, err := deployedPrimitiveStream.GetRecords(ctx, tsn_api.GetRecordsInput{
			DateFrom: unsafeParseDate("2020-01-01"),
			DateTo:   unsafeParseDate("2021-01-01"),
		})
		if !assert.NoError(t, err, "Failed to query records") {
			t.FailNow()
		}

		assert.Len(t, records, 1, "Expected exactly one record")
		assert.Equal(t, "1.000", records[0].Value.String(), "Unexpected record value")
		assert.Equal(t, "2020-01-01", records[0].DateValue.String(), "Unexpected record date")
	})
}

func unsafeParseDate(dateStr string) *civil.Date {
	date, err := civil.ParseDate(dateStr)
	if err != nil {
		panic(err)
	}
	return &date
}

func expectSuccessTx(t *testing.T, ctx context.Context, client client.Client, txHash transactions.TxHash) {
	txRes, err := client.WaitTx(ctx, txHash, time.Second)
	if !assert.NoError(t, err, "Transaction failed") {
		t.FailNow()
	}
	if !assert.Equal(t, transactions.CodeOk, transactions.TxCode(txRes.TxResult.Code), "Transaction code not OK: %s", txRes.TxResult.Log) {
		t.FailNow()
	}
}
