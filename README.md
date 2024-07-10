# Truflation Stream Network (TSN) SDK

The TSN SDK provides developers with tools to interact with the Truflation Stream Network, a decentralized platform for publishing, composing, and consuming economic data streams.

⚠ **These documentation files are a work in progress. They probably contain errors and inconsistencies. If you find some, don't hesitate to open an issue at the [GitHub repository](https://github.com/truflation/tsn-sdk).**

## Quick Start

### Prerequisites

- Go 1.20 or later

### Installation

```bash
go get github.com/truflation/tsn-sdk

```

### Example Usage

```go
package main

import (
	"context"
	"fmt"
	"github.com/golang-sql/civil"
	"github.com/truflation/tsn-sdk/core/tsnclient"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
)

func main() {
	ctx := context.Background()

	// Create TSN client
	tsnClient, err := tsnclient.NewClient(ctx, "<https://tsn-provider-url.com>")
	if err != nil {
		panic(err)
	}

	// Load an existing stream
	streamId := util.GenerateStreamId("your-stream-id")
	streamLocator := tsnClient.OwnStreamLocator(streamId)
	stream, err := tsnClient.LoadPrimitiveStream(streamLocator)
	if err != nil {
		panic(err)
	}

	// Read data from the stream
	records, err := stream.GetRecords(ctx, types.GetRecordsInput{
		DateFrom: civil.ParseDate("2023-01-01"),
		DateTo:   civil.ParseDate("2023-01-31"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(records)
}

```

## Types of Streams

- **Primitive Streams**: Direct data sources from providers. Examples include indexes from known sources, aggregation output such as sentiment analysis, and off-chain/on-chain data.
- **Composed Streams**: Aggregate and process data from multiple streams.
- **System Streams**: Contract-managed streams audited and accepted by TSN governance to ensure quality. See [type of streams guide](./docs/type-of-streams.md) for more information.

## Roles and Responsibilities

- **Data Providers**: Publish and maintain data streams, taxonomies, and push primitives.
- **Consumers**: Access and utilize stream data. Examples include researchers, analysts, financial institutions, and DApp developers.
- **Node Operators**: Maintain network infrastructure and consensus. Note: The network is currently in a centralized phase during development. Decentralization is planned for future releases. This repository does not handle node operation.

## Key Concepts

### Stream ID Composition

Stream IDs are unique identifiers generated for each stream. They ensure consistent referencing across the network.

### Types of Data Points

- **Record**: Raw data point provided by data sources.
- **Index**: Calculated values derived from stream data, representing a value's growth compared to the stream's first record.
- **Primitives**: Raw data points provided by data sources.

### Transaction Lifecycle

TSN operations rely on blockchain transactions. Some actions require waiting for previous transactions to be mined before proceeding. For detailed information on transaction dependencies and best practices, see [Stream Lifecycle](./docs/stream-lifecycle.md).

## Permissions and Privacy

TSN supports granular control over stream access and visibility. Streams can be public or private, with read and write permissions configurable at the wallet level. Additionally, you can control whether other streams can compose data from your stream. For more details, refer to [Stream Permissions](./docs/stream-permissions.md).

## Caveats

- **Transaction Confirmation**: Always wait for transaction confirmation before performing dependent actions. For more information, see the [Transaction Lifecycle](https://www.notion.so/Docs-561559c0d2344c3f92b14375f5b7eefe?pvs=21) section.

## Further Reading

- [Stream Lifecycle Documentation](./docs/stream-lifecycle.md)
- [Stream Permissions Guide](./docs/stream-permissions.md)
- [API Reference](./docs/api-reference.md)
- [Truflation Whitepaper](https://truflation.com/whitepaper)

For additional support or questions, please [open an issue](https://github.com/truflation/tsn-sdk/issues) or contact our support team.

[//]: # (TODO:)
[//]: # (- add "see tests for more examples" on the example usage)
[//]: # (- mention it uses kwil in the README)
[//]: # (- date input is wrong: we use civil date instead of string)
