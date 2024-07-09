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
    "github.com/truflation/tsn-sdk/core/client"
    "github.com/truflation/tsn-sdk/core/types"
)

func main() {
    ctx := context.Background()

    // Create TSN client
    tsnClient, err := client.NewClient(ctx, "<https://tsn-provider-url.com>")
    if err != nil {
        panic(err)
    }

    // Generate a stream ID
    streamId := util.GenerateStreamId("example-stream")

    // Deploy a new primitive stream
    deployTxHash, err := tsnClient.DeployStream(ctx, streamId, types.StreamTypePrimitive)
    if err != nil {
        panic(err)
    }
    fmt.Println("Deploy transaction hash:", deployTxHash)

    // Wait for the transaction to be mined
    txRes, err := tsnClient.WaitForTx(ctx, deployTxHash, time.Second)
    if err != nil {
        panic(err)
    }
    fmt.Println("Deploy transaction result:", txRes)

    // Load the deployed stream
    stream, err := tsnClient.LoadPrimitiveStream(tsnClient.OwnStreamLocator(streamId))
    if err != nil {
        panic(err)
    }

    // Initialize the stream
    txHashInit, err := stream.InitializeStream(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Println("Initialize transaction hash:", txHashInit)

    // Wait for the initialization transaction to be mined
    txResInit, err := tsnClient.WaitForTx(ctx, txHashInit, time.Second)
    if err != nil {
        panic(err)
    }
    fmt.Println("Initialize transaction result:", txResInit)

    // Insert records into the stream
    txHashInsert, err := stream.InsertRecords(ctx, []types.InsertRecordInput{
        {
            Value:     1,
            DateValue: civil.Date{Year: 2023, Month: 1, Day: 1},
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Insert transaction hash:", txHashInsert)

    // Wait for the insert transaction to be mined
    txResInsert, err := tsnClient.WaitForTx(ctx, txHashInsert, time.Second)
    if err != nil {
        panic(err)
    }
    fmt.Println("Insert transaction result:", txResInsert)

    // Read records from the stream
    records, err := stream.GetRecords(ctx, types.GetRecordsInput{
        DateFrom: civil.Date{Year: 2023, Month: 1, Day: 1},
        DateTo:   civil.Date{Year: 2023, Month: 12, Day: 31},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Records:", records)
}
```

Please refer to the test files in the SDK repository for more examples and detailed usage patterns. These tests provide comprehensive examples of various stream operations and error-handling scenarios.
