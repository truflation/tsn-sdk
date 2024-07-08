package integration

import (
	"context"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/truflation/tsn-sdk/internal/tsnclient"
	"github.com/truflation/tsn-sdk/internal/types"
	"github.com/truflation/tsn-sdk/internal/util"
	"testing"
	"time"
)

// ## Constants

const TestPrivateKey = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
const TestKwilProvider = "http://localhost:8484"

// ## Helper functions

// unsafeParseDate is a helper function to parse a date string into a civil.Date, panicking on error.
func unsafeParseDate(dateStr string) *civil.Date {
	date, err := civil.ParseDate(dateStr)
	if err != nil {
		panic(err)
	}
	return &date
}

// waitTxToBeMinedWithSuccess waits for a transaction to be successful, failing the test if it fails.
func waitTxToBeMinedWithSuccess(t *testing.T, ctx context.Context, client *tsnclient.Client, txHash transactions.TxHash) {
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
