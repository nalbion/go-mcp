package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// A uniquely identifying ID for a request in JSON-RPC.
type RequestId int

// A progress token, used to associate progress notifications with the original request.
type ProgressToken int

// A response to a request that indicates an error occurred.
type JSONRPCError struct {
	// Error corresponds to the JSON schema field "error".
	Error JSONRPCErrorError `json:"error" yaml:"error" mapstructure:"error"`

	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`
}

type JSONRPCErrorError struct {
	// The error type that occurred.
	Code int `json:"code" yaml:"code" mapstructure:"code"`

	// Additional information about the error. The value of this member is defined by
	// the sender (e.g. detailed error information, nested errors etc.).
	Data any `json:"data,omitempty" yaml:"data,omitempty" mapstructure:"data,omitempty"`

	// A short description of the error. The message SHOULD be limited to a concise
	// single sentence.
	Message string `json:"message" yaml:"message" mapstructure:"message"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONRPCErrorError) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["code"]; raw != nil && !ok {
		return fmt.Errorf("field code in JSONRPCErrorError: required")
	}
	if _, ok := raw["message"]; raw != nil && !ok {
		return fmt.Errorf("field message in JSONRPCErrorError: required")
	}
	type Plain JSONRPCErrorError
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = JSONRPCErrorError(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONRPCError) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["error"]; raw != nil && !ok {
		return fmt.Errorf("field error in JSONRPCError: required")
	}
	if _, ok := raw["id"]; raw != nil && !ok {
		return fmt.Errorf("field id in JSONRPCError: required")
	}
	if _, ok := raw["jsonrpc"]; raw != nil && !ok {
		return fmt.Errorf("field jsonrpc in JSONRPCError: required")
	}
	type Plain JSONRPCError
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = JSONRPCError(plain)
	return nil
}

type JSONRPCMessage any

// A notification which does not expect a response.
type JSONRPCNotification struct {
	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *JSONRPCNotificationParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type JSONRPCNotificationParams struct {
	// This parameter name is reserved by MCP to allow clients and servers to attach
	// additional metadata to their notifications.
	Meta *JSONRPCNotificationParamsMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	AdditionalProperties any `mapstructure:",remain"`
}

// This parameter name is reserved by MCP to allow clients and servers to attach
// additional metadata to their notifications.
type JSONRPCNotificationParamsMeta map[string]any

// MarshalJSON implements json.Marshaler.
func (j *JSONRPCNotificationParams) MarshalJSON() ([]byte, error) {
	return marshalParam(j.AdditionalProperties, (*map[string]any)(j.Meta))
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONRPCNotification) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["jsonrpc"]; raw != nil && !ok {
		return fmt.Errorf("field jsonrpc in JSONRPCNotification: required")
	}
	if _, ok := raw["method"]; raw != nil && !ok {
		return fmt.Errorf("field method in JSONRPCNotification: required")
	}
	type Plain JSONRPCNotification
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = JSONRPCNotification(plain)
	return nil
}

// A request that expects a response.
type JSONRPCRequest struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	Params *JSONRPCRequestParams `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type JSONRPCRequestParams struct {
	// Meta corresponds to the JSON schema field "_meta".
	Meta *JSONRPCRequestParamsMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	AdditionalProperties any `json:",inline" mapstructure:",remain"`
}

type JSONRPCRequestParamsMeta map[string]any

// MarshalJSON implements json.Marshaler.
func (j *JSONRPCRequestParams) MarshalJSON() ([]byte, error) {
	return marshalParam(j.AdditionalProperties, (*map[string]any)(j.Meta))
}

func marshalParam(params any, meta *map[string]any) ([]byte, error) {
	baseMap := map[string]any{}
	if meta != nil {
		baseMap["_meta"] = meta
	}

	if params != nil {
		additionalProperties, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		var additionalPropertiesMap map[string]any
		if err := json.Unmarshal(additionalProperties, &additionalPropertiesMap); err != nil {
			return nil, err
		}
		for k, v := range additionalPropertiesMap {
			baseMap[k] = v
		}
	}
	return json.Marshal(baseMap)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONRPCRequest) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["id"]; raw != nil && !ok {
		return fmt.Errorf("field id in JSONRPCRequest: required")
	}
	if _, ok := raw["jsonrpc"]; raw != nil && !ok {
		return fmt.Errorf("field jsonrpc in JSONRPCRequest: required")
	}
	if _, ok := raw["method"]; raw != nil && !ok {
		return fmt.Errorf("field method in JSONRPCRequest: required")
	}
	type Plain JSONRPCRequest
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = JSONRPCRequest(plain)
	return nil
}

// type Result any
type Result struct {
	// This result property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their responses.
	Meta ResultMeta `json:"_meta,omitempty" yaml:"_meta,omitempty" mapstructure:"_meta,omitempty"`

	AdditionalProperties any `mapstructure:",remain"`
}

type ResultMeta map[string]any

// A successful (non-error) response to a request.
type JSONRPCResponse struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Result corresponds to the JSON schema field "result".
	Result Result `json:"result" yaml:"result" mapstructure:"result"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONRPCResponse) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["id"]; raw != nil && !ok {
		return fmt.Errorf("field id in JSONRPCResponse: required")
	}
	if _, ok := raw["jsonrpc"]; raw != nil && !ok {
		return fmt.Errorf("field jsonrpc in JSONRPCResponse: required")
	}
	if _, ok := raw["result"]; raw != nil && !ok {
		return fmt.Errorf("field result in JSONRPCResponse: required")
	}
	type Plain JSONRPCResponse
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = JSONRPCResponse(plain)
	return nil
}
