# API Reference

The Truf Node SDK offers a comprehensive set of APIs for interacting with the TN, allowing developers to create, manage, and consume data streams. This document provides detailed descriptions of all available methods in the SDK, along with examples of their usage.

## Interfaces
- [Client](client.md)
- [Primitive Stream](primitive-stream.md)
- [Composed Stream](composed-stream.md)

Other utilities are in the [util](util.md) documentation.

## Example Usage

Below is an example demonstrating how to use the TN SDK to deploy, initialize, and read from a primitive stream.

```go
package main

import (
	"context"
	"fmt"
	"github.com/golang-sql/civil"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/trufnetwork/truf-node-sdk-go/core/tnclient"
	"github.com/trufnetwork/truf-node-sdk-go/core/types"
	"github.com/trufnetwork/truf-node-sdk-go/core/util"
	"time"
)

func main() {
	// handle errors appropriately in a real application
	ctx := context.Background()

	// Create TN client
	pk, _ := crypto.Secp256k1PrivateKeyFromHex("<your-private-key-hex>")
	signer := &auth.EthPersonalSigner{Key: *pk}
	tnClient, _ := tnclient.NewClient(ctx, "<https://tsn-provider-url.com>", tnclient.WithSigner(signer))

	// Generate a stream ID
	streamId := util.GenerateStreamId("example-stream")

	// Deploy a new primitive stream
	deployTxHash, _ := tnClient.DeployStream(ctx, streamId, types.StreamTypePrimitive)

	// Wait for the transaction to be mined
	txRes, _ := tnClient.WaitForTx(ctx, deployTxHash, time.Second)

	// Load the deployed stream
	stream, _ := tnClient.LoadPrimitiveStream(tnClient.OwnStreamLocator(streamId))

	// Initialize the stream
	txHashInit, _ := stream.InitializeStream(ctx)

	// Wait for the initialization transaction to be mined
	txResInit, _ := tnClient.WaitForTx(ctx, txHashInit, time.Second)
	fmt.Println("Initialize transaction result:", txResInit)

	// Insert records into the stream
	txHashInsert, _ := stream.InsertRecords(ctx, []types.InsertRecordInput{
		{
			Value:     1,
			DateValue: civil.Date{Year: 2023, Month: 1, Day: 1},
		},
	})

	// Wait for the insert transaction to be mined
	_, _ = tnClient.WaitForTx(ctx, txHashInsert, time.Second)

	// Read records from the stream
	records, _ := stream.GetRecord(ctx, types.GetRecordInput{
		DateFrom: civil.ParseDate("2023-01-01"),
		DateTo:   civil.ParseDate("2023-01-31"),
	})
	fmt.Println("Records:", records)
}

```

Please refer to the test files in the SDK repository for more examples and detailed usage patterns. These tests provide comprehensive examples of various stream operations and error-handling scenarios.
