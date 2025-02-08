# Model Context Protocol

The code here was adapted from the [kotlin-sdk](https://github.com/modelcontextprotocol/kotlin-sdk) and [typescript-sdk](https://github.com/modelcontextprotocol/typescript-sdk).

See also https://modelcontextprotocol.io/introduction

## MCP Client

When creating a `Client`, list it's capabilities, and then `client.Connect(transport)` connects to the `Transport` (through `Protocol` which it extends) and starts the `initialize` process.

You can then call `client.ListTools()`, `CallTool()` etc.

## MCP Server

Creating a `Server` is similar to creating a `Client` - define its capabilities and call `server.Connect(transport)` and it will start listening for messages through the `Transport`.

# JSON RPC

The `Transport` classes and `Protocol` are independant of MCP and could also be used for LSP client/servers etc.

## JSON RCP Transport

The `Transport` class and it's `client`/`server` implementations for `stdio`, `sse`, `in_memory` know nothing about MCP. The role of the `Transport` interface is to simply send/receive JSON RPC messages, it does not format/parse messages, that's the role of `Protocol`.

## JSON RPC Protocol

The `Protocol` class is provided for formatting and parsing JSON into Request/Response/Notification/Error messages and delegates the send/receive to `Transport`.
