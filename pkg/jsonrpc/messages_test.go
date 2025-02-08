package jsonrpc

import (
	"encoding/json"
	"testing"

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

	var result JSONRPCMessage
	err := json.Unmarshal(jsonResult, &result)
	require.NoError(t, err)

	expectedResult := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"logging": map[string]interface{}{},
			"prompts": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"subscribe":   true,
				"listChanged": true,
			},
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "ExampleServer",
			"version": "1.0.0",
		},
	}

	t.Run("nil responseMessage", func(t *testing.T) {
		// when
		err := ParseResult(jsonResult, nil)

		// then
		require.Error(t, err)
	})

	t.Run("empty responseMessage", func(t *testing.T) {
		// when
		messageResult := &Result{}
		err := ParseResult(jsonResult, messageResult)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedResult, messageResult.AdditionalProperties)
	})
}

func TestParseJSONRPCMessage(t *testing.T) {
	t.Run("request message", func(t *testing.T) {
		// given
		jsonRequest := []byte(`{
	  "jsonrpc": "2.0",
	  "id": 1,
	  "method": "initialize",
	  "params": {
	    "protocolVersion": "2024-11-05",
	    "capabilities": {
	      "roots": {
	        "listChanged": true
	      },
	      "sampling": {}
	    },
	    "clientInfo": {
	      "name": "ExampleClient",
	      "version": "1.0.0"
	    }
	  }
	}`)

		expectedRequest := JSONRPCRequest{
			Jsonrpc: "2.0",
			Id:      1,
			Method:  "initialize",
			Params: &JSONRPCRequestParams{
				AdditionalProperties: map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities": map[string]interface{}{
						"roots": map[string]interface{}{
							"listChanged": true,
						},
						"sampling": map[string]interface{}{},
					},
					"clientInfo": map[string]interface{}{
						"name":    "ExampleClient",
						"version": "1.0.0",
					},
				},
			},
		}

		message, err := ParseJSONRPCMessage(jsonRequest)
		require.IsType(t, message, &JSONRPCRequest{})

		// then
		require.NoError(t, err)
		require.Equal(t, expectedRequest, *message.(*JSONRPCRequest))
	})

	t.Run("notification message", func(t *testing.T) {
		// given
		jsonNotification := []byte(`{
		"jsonrpc": "2.0",
		"method": "notifications/cancelled",
	    "params": {
	      "requestId": 123,
	      "reason": "User requested cancellation"
	    }
	  }`)

		reason := "User requested cancellation"
		expectedNotification := JSONRPCNotification{
			Jsonrpc: "2.0",
			Method:  "notifications/cancelled",
			Params: &JSONRPCNotificationParams{
				AdditionalProperties: map[string]interface{}{
					"requestId": float64(123), // TODO: should be int
					"reason":    reason,
				},
			},
		}

		notification, err := ParseJSONRPCMessage(jsonNotification)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedNotification, *notification.(*JSONRPCNotification))
	})

	t.Run("error message", func(t *testing.T) {
		// given
		jsonResponse := []byte(`{
	"jsonrpc": "2.0",
	"id": 1,
	"error": {
		"code": -32602,
		"message": "Unsupported protocol version",
		"data": {
			"supported": ["2024-11-05"],
			"requested": "1.0.0"
		}
	}
}`)

		expectedError := JSONRPCError{
			Jsonrpc: "2.0",
			Id:      1,
			Error: JSONRPCErrorError{
				Code:    -32602,
				Message: "Unsupported protocol version",
				Data: map[string]interface{}{
					"supported": []interface{}{"2024-11-05"},
					"requested": "1.0.0",
				},
			},
		}

		result, err := ParseJSONRPCMessage(jsonResponse)

		// then
		require.NoError(t, err)
		require.Equal(t, expectedError, result)
	})
}
