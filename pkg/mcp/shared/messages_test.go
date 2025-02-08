package shared

import (
	"encoding/json"
	"testing"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/stretchr/testify/require"
)

func TestParseResult(t *testing.T) {
	// given
	jsonResult := []byte(`{
  "protocolVersion": "2024-11-05",
  "capabilities": {
    "logging": {},
    "prompts": {
      "listChanged": true
    },
    "resources": {
      "subscribe": true,
      "listChanged": true
    },
    "tools": {
      "listChanged": true
    }
  },
  "serverInfo": {
    "name": "ExampleServer",
    "version": "1.0.0"
  }
}`)

	trueVal := true
	var result jsonrpc.JSONRPCMessage
	err := json.Unmarshal(jsonResult, &result)
	require.NoError(t, err)

	expectedResult := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Logging: ServerCapabilitiesLogging{},
			Prompts: &ServerCapabilitiesPrompts{
				ListChanged: &trueVal,
			},
			Resources: &ServerCapabilitiesResources{
				Subscribe:   &trueVal,
				ListChanged: &trueVal,
			},
			Tools: &ServerCapabilitiesTools{
				ListChanged: &trueVal,
			},
		},
		ServerInfo: Implementation{
			Name:    "ExampleServer",
			Version: "1.0.0",
		},
	}

	t.Run("nil responseMessage", func(t *testing.T) {
		// when
		err := jsonrpc.ParseResult(jsonResult, nil)

		// then
		require.Error(t, err)
	})

	t.Run("empty responseMessage", func(t *testing.T) {
		// when
		messageResult := jsonrpc.Result{
			AdditionalProperties: InitializeResult{},
		}
		err := jsonrpc.ParseResult(jsonResult, &messageResult)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedResult, messageResult.AdditionalProperties)
	})
}

func TestParseJSONRPCMessage(t *testing.T) {
// 	t.Run("request message", func(t *testing.T) {
// 		// given
// 		jsonRequest := []byte(`{
//   "jsonrpc": "2.0",
//   "id": 1,
//   "method": "initialize",
//   "params": {
//     "protocolVersion": "2024-11-05",
//     "capabilities": {
//       "roots": {
//         "listChanged": true
//       },
//       "sampling": {}
//     },
//     "clientInfo": {
//       "name": "ExampleClient",
//       "version": "1.0.0"
//     }
//   }
// }`)

// 		trueVal := true
// 		expectedRequest := jsonrpc.JSONRPCRequest{
// 			Jsonrpc: "2.0",
// 			Id:      1,
// 			Method:  "initialize",
// 			Params: &jsonrpc.JSONRPCRequestParams{
// 				// JSONRPCRequestParamsMeta only allows ProgressToken?
// 				// Meta: map[string]interface{}{
// 				// 	"foo": "bar",
// 				// },
// 				AdditionalProperties: InitializeRequestParams{
// 					ProtocolVersion: "2024-11-05",
// 					Capabilities: ClientCapabilities{
// 						Roots: &ClientCapabilitiesRoots{
// 							ListChanged: &trueVal,
// 						},
// 						Sampling: ClientCapabilitiesSampling{},
// 					},
// 					ClientInfo: Implementation{
// 						Name:    "ExampleClient",
// 						Version: "1.0.0",
// 					},
// 				},
// 			},
// 		}

// 		message, err := jsonrpc.ParseJSONRPCMessage(jsonRequest)
// 		require.IsType(t, message, &jsonrpc.JSONRPCRequest{})

// 		// then
// 		require.NoError(t, err)
// 		require.Equal(t, expectedRequest, *message.(*jsonrpc.JSONRPCRequest))
// 	})

	// 	t.Run("response message", func(t *testing.T) {
	// 		// given
	// 		jsonResponse := []byte(`{
	// 	"jsonrpc": "2.0",
	// 	"id": 1,
	// 	"result": {
	// 		"_meta": {
	// 		   "foo": "bar"
	// 		},
	// 		"protocolVersion": "2024-11-05",
	// 		"capabilities": {
	// 			"logging": {},
	// 			"prompts": {
	// 				"listChanged": true
	// 			},
	// 			"resources": {
	// 				"subscribe": true,
	// 				"listChanged": true
	// 			},
	// 			"tools": {
	// 				"listChanged": true
	// 			}
	// 		},
	// 		"serverInfo": {
	// 			"name": "ExampleServer",
	// 			"version": "1.0.0"
	// 		}
	// 	}
	// }`)

	// 		trueVal := true
	// 		expectedResult := JSONRPCResponse{
	// 			Jsonrpc: "2.0",
	// 			Id:      1,
	// 			Result: Result{
	// 				AdditionalProperties: InitializeResult{
	// 					ProtocolVersion: "2024-11-05",
	// 					Capabilities: ServerCapabilities{
	// 						Logging: ServerCapabilitiesLogging{},
	// 						Prompts: &ServerCapabilitiesPrompts{
	// 							ListChanged: &trueVal,
	// 						},
	// 						Resources: &ServerCapabilitiesResources{
	// 							Subscribe:   &trueVal,
	// 							ListChanged: &trueVal,
	// 						},
	// 						Tools: &ServerCapabilitiesTools{
	// 							ListChanged: &trueVal,
	// 						},
	// 					},
	// 					ServerInfo: Implementation{
	// 						Name:    "ExampleServer",
	// 						Version: "1.0.0",
	// 					},
	// 				},
	// 			},
	// 		}

	// 		message, err := ParseJSONRPCMessage(jsonResponse)
	// 		require.IsType(t, message, JSONRPCResponse{})

	// 		// then
	// 		require.NoError(t, err)
	// 		require.Equal(t, expectedResult, message)
	// 		require.Equal(t, "bar", message.(JSONRPCResponse).Result.Meta["foo"])
	// 	})

// 	t.Run("notification message", func(t *testing.T) {
// 		// given
// 		jsonNotification := []byte(`{
// 	"jsonrpc": "2.0",
// 	"method": "notifications/cancelled",
//     "params": {
//       "requestId": 123,
//       "reason": "User requested cancellation"
//     }
//   }`)

// 		reason := "User requested cancellation"
// 		expectedNotification := jsonrpc.JSONRPCNotification{
// 			Jsonrpc: "2.0",
// 			Method:  "notifications/cancelled",
// 			Params: &jsonrpc.JSONRPCNotificationParams{
// 				AdditionalProperties: map[string]interface{}{
// 					"requestId": float64(123),
// 					"reason":    reason,
// 				},
// 			},
// 		}

// 		notification, err := jsonrpc.ParseJSONRPCMessage(jsonNotification)

// 		// then
// 		require.NoError(t, err)
// 		require.Equal(t, expectedNotification, *notification.(*jsonrpc.JSONRPCNotification))
// 	})
}
