package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

	meta, params, err := parseAdditionalProperties(content)
	if err != nil {
		return nil, err
	}

	metaParams := JSONRPCNotificationParamsMeta(meta)
	if metaParams != nil {
		request.Params.Meta = &metaParams
	}
	request.Params.AdditionalProperties = params

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

	if request.Params != nil {
		meta, params, err := parseAdditionalProperties(content)
		if err != nil {
			return nil, err
		}

		metaParams := JSONRPCRequestParamsMeta(meta)
		if metaParams != nil {
			request.Params.Meta = &metaParams
		}
		request.Params.AdditionalProperties = params
	}

	return &request, nil
}

func parseAdditionalProperties(content []byte) (map[string]any, map[string]any, error) {
	jsonRequest := map[string]interface{}{}
	err := json.Unmarshal(content, &jsonRequest)
	if err != nil {
		return nil, nil, err
	}

	var meta map[string]any
	var params map[string]any

	if metaValue, ok := jsonRequest["_meta"]; ok {
		if meta, ok = metaValue.(map[string]interface{}); !ok {
			return nil, nil, errors.New("invalid _meta")
		}
	}

	if paramsValue, ok := jsonRequest["params"]; ok {
		if params, ok = paramsValue.(map[string]interface{}); !ok {
			return nil, nil, errors.New("invalid params")
		}
	}

	return meta, params, nil
}

func ParseResult(content []byte, messageResult *Result) error {
	if messageResult == nil {
		return errors.New("messageResult is nil")
	}

	if err := json.Unmarshal(content, messageResult); err != nil {
		return err
	}

	if messageResult.AdditionalProperties == nil {
		messageResult.AdditionalProperties = make(map[string]interface{})
	}

	resultType := reflect.TypeOf(messageResult.AdditionalProperties)
	isPointer := resultType.Kind() == reflect.Ptr

	if isPointer {
		resultType = resultType.Elem()
	}

	resultValue := reflect.New(resultType).Interface()

	if err := json.Unmarshal(content, resultValue); err != nil {
		return err
	}

	if isPointer {
		messageResult.AdditionalProperties = resultValue
	} else {
		messageResult.AdditionalProperties = reflect.ValueOf(resultValue).Elem().Interface()
	}

	return nil
}

type JSONRPCNotificationMessage interface{}

// func (e *JSONRPCErrorError) Error() string {
// 	return fmt.Sprintf("code: %d, message: %s, data: %v", e.Code, e.Message, e.Data)
// }
