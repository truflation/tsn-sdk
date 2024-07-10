# Primitive Stream Interface

The `IPrimitiveStream` interface extends `IStream` and provides additional methods for interacting with primitive streams.

## Methods

### `InsertRecords`

Inserts records into the stream.

```go
InsertRecords(ctx context.Context, inputs []types.InsertRecordInput) (transactions.TxHash, error)
```

**Parameters:**
- `ctx`: The context for the operation.
- `inputs`: A slice of `InsertRecordInput` representing the records to be inserted.

**Returns:**
- `transactions.TxHash`: The transaction hash for the operation.
- `error`: An error if the operation fails.
```