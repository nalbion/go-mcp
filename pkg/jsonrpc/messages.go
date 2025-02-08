package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Method string

func ParseJSONRPCMessage(content []byte) (JSONRPCMessage, error) {
	parsed := map[string]interface{}{}
	err := json.Unmarshal(content, &parsed)
	if err != nil {
		return nil, err
	}

	// could be any of:
	// - JSONRPCRequest:      jsonrpc, id, method, [params]
	// - JSONRPCResponse:     jsonrpc, id, result
	// - JSONRPCNotification: jsonrpc,     method, [params]
	// - JSONRPCError:        jsonrpc, id, error
	if _, ok := parsed["method"]; ok {
		if _, ok := parsed["id"]; ok {
			return parseRequest(content)
		} else {
			return parseNofication(content)
		}
	} else if _, ok := parsed["result"]; ok {
		var message JSONRPCResponse
		if err = json.Unmarshal(content, &message); err != nil {
			return nil, err
		}
		// if message.Result.AdditionalProperties == nil {
		// 	// probably a bug in the generated mcp_models code
		// 	messageMap := map[string]interface{}{}
		// 	if err = json.Unmarshal(content, &messageMap); err != nil {
		// 		return nil, err
		// 	}

		// 	if result, ok := messageMap["result"].(map[string]interface{}); ok {
		// 		// delete(result, "_meta")

		// 		if marshalledResult, err := json.Marshal(result); err != nil {
		// 			return nil, err
		// 		} else {
		// 			if message.Result.AdditionalProperties, err = parseResult(marshalledResult); err != nil {
		// 				return nil, err
		// 			}
		// 		}
		// 	}
		// }

		return message, nil
	} else if _, ok := parsed["error"]; ok {
		var message JSONRPCError
		if err = json.Unmarshal(content, &message); err != nil {
			return nil, err
		}
		return message, nil
	} else {
		return nil, fmt.Errorf("unknown message type: %s", content)
	}
}

func parseNofication(content []byte) (*JSONRPCNotification, error) {
	var request JSONRPCNotification
	err := json.Unmarshal(content, &request)
	if err != nil {
		return nil, err
	}

	if request.Method == "" {
		return nil, errors.New("no method provided")
	}

	// switch Method(request.Method) {
	// case NotificationsCancelledMethod:
	// 	var message CancelledNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case NotificationsInitializedMethod:
	// 	var message InitializedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case NotificationsProgressMethod:
	// 	var message ProgressNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case LoggingMessageNotificationMethod:
	// 	var message LoggingMessageNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ResourceUpdatedNotificationMethod:
	// 	var message ResourceUpdatedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ResourceListChangedNotificationMethod:
	// 	var message ResourceListChangedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ToolListChangedNotificationMethod:
	// 	var message ToolListChangedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case NotificationsRootsListChangedMethod:
	// 	var message RootsListChangedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case NotificationsPromptListChangedMethod:
	// 	var message PromptListChangedNotification
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// default:
	// 	return nil, fmt.Errorf("unknown method: %s", request.Method)
	// }

	return &request, nil
}

func parseRequest(content []byte) (*JSONRPCRequest, error) {
	var request JSONRPCRequest
	err := json.Unmarshal(content, &request)
	if err != nil {
		return nil, err
	}

	if request.Method == "" {
		return nil, errors.New("no method provided")
	}

	// switch Method(request.Method) {
	// case InitializeMethod:
	// 	var message InitializeRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case PingMethod:
	// 	var message PingRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ListResourcesMethod:
	// 	var message ListResourcesRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ListResourcesTemplatesMethod:
	// 	var message ListResourceTemplatesRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ReadResourcesMethod:
	// 	var message ReadResourceRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ResourcesSubscribeMethod:
	// 	var message SubscribeRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ResourcesUnsubscribeMethod:
	// 	var message UnsubscribeRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ListPromptsMethod:
	// 	var message ListPromptsRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case GetPromptsMethod:
	// 	var message GetPromptRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ToolsListMethod:
	// 	var message ListToolsRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case ToolsCallMethod:
	// 	var message CallToolRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case LoggingSetLevelMethod:
	// 	var message SetLevelRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case SamplingCreateMessageMethod:
	// 	var message CreateMessageRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case CompletionCompleteMethod:
	// 	var message CompleteRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// case RootsListMethod:
	// 	var message ListRootsRequest
	// 	if err = json.Unmarshal(content, &message); err != nil {
	// 		return nil, err
	// 	}
	// 	request.Method = message.Method
	// 	request.Params.AdditionalProperties = message.Params
	// default:
	// 	return nil, fmt.Errorf("unknown method: %s", request.Method)
	// }

	return &request, nil
}

func ParseResult(content []byte, messageResult *Result) error {
	if messageResult == nil {
		return errors.New("messageResult is nil")
	}

	if err := json.Unmarshal(content, messageResult); err != nil {
		return err
	}
	return nil
}

type JSONRPCNotificationMessage interface{}

// func (e *JSONRPCErrorError) Error() string {
// 	return fmt.Sprintf("code: %d, message: %s, data: %v", e.Code, e.Message, e.Data)
// }
