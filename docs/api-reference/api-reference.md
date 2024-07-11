# API Reference

The Truflation Stream Network (TSN) SDK offers a comprehensive set of APIs for interacting with the TSN, allowing developers to create, manage, and consume data streams. This document provides detailed descriptions of all available methods in the SDK, along with examples of their usage.

## Interfaces
- [Client](client.md)
- [Primitive Stream](primitive-stream.md)
- [Composed Stream](composed-stream.md)

Other utilities are in the [util](util.md) documentation.

## Example Usage

Below is an example demonstrating how to use the TSN SDK to deploy, initialize, and read from a primitive stream.

```go
package main

import (
	"context"
	"fmt"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
	"time"
)

func main() {
	// handle errors appropriately in a real application
	ctx := context.Background()

	// Create TSN client
	pk, _ := crypto.Secp256k1PrivateKeyFromHex("<your-private-key-hex>")
	signer := &auth.EthPersonalSigner{Key: *pk}
	tsnClient, _ := tsnclient.NewClient(ctx, "<https://tsn-provider-url.com>", tsnclient.WithSigner(signer))

	// Generate a stream ID
	streamId := util.GenerateStreamId("example-stream")

	// Deploy a new primitive stream
	deployTxHash, _ := tsnClient.DeployStream(ctx, streamId, types.StreamTypePrimitive)

	// Wait for the transaction to be mined
	txRes, _ := tsnClient.WaitForTx(ctx, deployTxHash, time.Second)

	// Load the deployed stream
	stream, _ := tsnClient.LoadPrimitiveStream(tsnClient.OwnStreamLocator(streamId))

	// Initialize the stream
	txHashInit, _ := stream.InitializeStream(ctx)

	// Wait for the initialization transaction to be mined
	txResInit, _ := tsnClient.WaitForTx(ctx, txHashInit, time.Second)
	fmt.Println("Initialize transaction result:", txResInit)

	// Insert records into the stream
	txHashInsert, _ := stream.InsertRecords(ctx, []types.InsertRecordInput{
		{
			Value:     1,
			DateValue: civil.Date{Year: 2023, Month: 1, Day: 1},
		},
	})

	// Wait for the insert transaction to be mined
	_, _ = tsnClient.WaitForTx(ctx, txHashInsert, time.Second)

	// Read records from the stream
	records, _ := stream.GetRecord(ctx, types.GetRecordInput{
		DateFrom: civil.ParseDate("2023-01-01"),
		DateTo:   civil.ParseDate("2023-01-31"),
	})
	fmt.Println("Records:", records)
}

```

Please refer to the test files in the SDK repository for more examples and detailed usage patterns. These tests provide comprehensive examples of various stream operations and error-handling scenarios.
