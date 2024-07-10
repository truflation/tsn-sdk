# Default Contracts in TSN

The Truflation Stream Network (TSN) uses [kuneiform data contracts](https://docs.kwil.com/docs/kuneiform/introduction) to manage data streams. This document explains the default contracts for primitive and composed streams, their customization potential, and key features.

You can find the default contracts in the `core/contracts` directory of the TSN SDK.

## Contract Customization

- Data providers can modify the logic of primitive and composed stream contracts to implement different methodologies for data calculation and composition.
- The interfaces between contracts (e.g., `get_index`, `get_record`) and the permissions logic must remain unchanged to ensure compatibility within the TSN ecosystem.
- While the TSN SDK is not yet fully prepared for contract modifications, users with the necessary expertise can implement custom logic. Future updates to the SDK will enhance support for customized contracts.

## Metadata Management

All contracts include a metadata table with keys to manage:
- Permissions (e.g., `read_visibility`, `compose_visibility`, `allow_read_wallet`, `allow_compose_stream`)
- Ownership (e.g., `stream_owner`)
- Other stream-specific information

## Primitive Stream Default Contract

The primitive stream contract is designed to store raw data and calculate indexes:

- Stores date-value pairs for the stream
- Provides functions to insert and retrieve records
- Calculates indexes based on the stream's base value
- Implements permission checks for read and write access

Key procedures:
- `insert_record`: Allows data providers to add new records
- `get_record`: Retrieves data, filling gaps with the last known value
- `get_index`: Calculates the index based on the base value of the stream

This basic functionality is typically sufficient for most use cases and may not require modification.

## Composed Stream Default Contract

The composed stream contract aggregates data from multiple child streams:

- Stores a taxonomy of child streams and their associated weights
- Uses weights to calculate a weighted average of child stream values
- Provides functions to manage the taxonomy and retrieve composed data

Key procedures:
- `set_taxonomy`: Manages the structure and weights of child streams
- `get_record`: Retrieves weighted average data from child streams
- `get_index`: Calculates a weighted index based on child stream indexes

While this default logic suits many scenarios, data providers may customize it to implement more complex composition strategies.

## Future Customization Support

The TSN team is working on enhancing the SDK to better support contract customization. This will enable data providers to more easily implement unique data methodologies while maintaining compatibility with the TSN ecosystem.

Until then, advanced users can modify contract logic directly, ensuring they maintain the required interfaces and permission structures.