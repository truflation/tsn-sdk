# Stream Interface

The `IStream` and provides methods for interacting with any stream. It's the common interface for both primitive and composed streams.

## Methods

### `InitializeStream`

```go
InitializeStream(ctx context.Context) (transactions.TxHash, error)
```

Initializes the primitive stream.

**Parameters:**
- `ctx`: The context for the operation.

**Returns:**
- `transactions.TxHash`: The transaction hash for the initialization.
- `error`: An error if the initialization fails.

### `GetRecord`

```go
GetRecord(ctx context.Context, input types.GetRecordInput) ([]types.StreamRecord, error)
```

Retrieves records from the stream based on the input criteria.

**Parameters:**
- `ctx`: The context for the operation.
- `input`: The input criteria for retrieving records.

**Returns:**
- `[]types.StreamRecord`: The retrieved records.
- `error`: An error if the retrieval fails.

### `SetReadVisibility`

```go
SetReadVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
```

Sets the read visibility of the stream.

**Parameters:**
- `ctx`: The context for the operation.
- `visibility`: The visibility setting (`Public`, `Private`).

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

### `SetComposeVisibility`

```go
SetComposeVisibility(ctx context.Context, visibility util.VisibilityEnum) (transactions.TxHash, error)
```

Sets the compose visibility of the stream.

**Parameters:**
- `ctx`: The context for the operation.
- `visibility`: The visibility setting (`Public`, `Private`).

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

### `AllowReadWallet`

```go
AllowReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
```

Allows a wallet to read the stream.

**Parameters:**
- `ctx`: The context for the operation.
- `wallet`: The Ethereum address of the wallet.

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

### `DisableReadWallet`

```go
DisableReadWallet(ctx context.Context, wallet util.EthereumAddress) (transactions.TxHash, error)
```

Disables a wallet from reading the stream.

**Parameters:**
- `ctx`: The context for the operation.
- `wallet`: The Ethereum address of the wallet.

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

### `AllowComposeStream`

```go
AllowComposeStream(ctx context.Context, locator StreamLocator) (transactions.TxHash, error)
```

Allows a stream to use this stream as a child.

**Parameters:**
- `ctx`: The context for the operation.
- `locator`: The locator of the composed stream.

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

### `DisableComposeStream`

```go
DisableComposeStream(ctx context.Context, locator StreamLocator) (transactions.TxHash, error)
```

Disables a stream from using this stream as a child.

**Parameters:**
- `ctx`: The context for the operation.
- `locator`: The locator of the composed stream.

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.

