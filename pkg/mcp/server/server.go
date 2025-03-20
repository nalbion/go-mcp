package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/nalbion/go-mcp/pkg/jsonrpc"
	"github.com/nalbion/go-mcp/pkg/mcp"
	"github.com/nalbion/go-mcp/pkg/mcp/shared"
)

type ServerOptions struct {
	// ProtocolOptions contains common protocol options
	shared.ProtocolOptions
	// Capabilities defines the capabilities this server supports
	Capabilities mcp.ServerCapabilities
	// Instructions provides optional instructions to clients
	Instructions string
	Logger       shared.MCPLogger
}

func NewServerOptions() ServerOptions {
	return ServerOptions{
		ProtocolOptions: shared.ProtocolOptions{
			EnforceStrictCapabilities: true,
		},
		Capabilities: mcp.ServerCapabilities{},
		Logger:       shared.DefaultLogger,
	}
}

// An MCP server on top of a pluggable transport.
// This server automatically responds to the initialization flow as initiated by the client.
// You can register tools, prompts, and resources using AddTool(), AddPrompt()), and AddResource().
// The server will then automatically handle listing and retrieval requests from the client.
type Server struct {
	*shared.Protocol
	ctx                context.Context
	serverInfo         mcp.Implementation
	options            *ServerOptions
	clientCapabilities *mcp.ClientCapabilities
	clientVersion      *mcp.Implementation
	capabilities       mcp.ServerCapabilities
	instructions       string

	tools             map[string]RegisteredTool
	prompts           map[string]RegisteredPrompt
	resources         map[string]RegisteredResource
	resourceTemplates map[string]RegisteredResourceTemplate

	onInitialized jsonrpc.NotificationHandler
	onClose       func()

	toolsMutex     sync.RWMutex
	promptsMutex   sync.RWMutex
	resourcesMutex sync.RWMutex
	templatesMutex sync.RWMutex

	logger shared.MCPLogger
}

func NewServer(ctx context.Context, serverInfo mcp.Implementation, options *ServerOptions) *Server {
	s := &Server{
		Protocol:          shared.NewProtocol(ctx, &options.ProtocolOptions),
		ctx:               ctx,
		serverInfo:        serverInfo,
		options:           options,
		capabilities:      options.Capabilities,
		instructions:      options.Instructions,
		tools:             make(map[string]RegisteredTool),
		prompts:           make(map[string]RegisteredPrompt),
		resources:         make(map[string]RegisteredResource),
		resourceTemplates: make(map[string]RegisteredResourceTemplate),
		onClose:           func() {},
		logger:            options.Logger,
	}

	// If no logger was provided, use the default logger
	if s.logger == nil {
		s.logger = shared.DefaultLogger
	}

	s.logger.Info("Initializing MCP server with capabilities: %v", s.capabilities)

	s.SetContext(s.ctx)

	s.SetRequestHandler(shared.InitializeMethod, s.handleInitialize)
	s.SetNotificationHandler(shared.InitializedMethod, s.onInitialized)

	if s.capabilities.Tools != nil {
		s.SetRequestHandler(shared.ToolsListMethod, s.HandleListTools)
		s.SetRequestHandler(shared.ToolsCallMethod, s.HandleCallTool)
	}

	if s.capabilities.Prompts != nil {
		s.SetRequestHandler(shared.ListPromptsMethod, s.handleListPrompts)
		s.SetRequestHandler(shared.GetPromptsMethod, s.handleGetPrompt)
	}

	if s.capabilities.Resources != nil {
		s.SetRequestHandler(shared.ListResourcesMethod, s.handleListResources)
		s.SetRequestHandler(shared.ReadResourcesMethod, s.handleReadResource)
		// s.SetRequestHandler(shared.ListResourcesTemplatesMethod, s.handleListResourceTemplates)
	}

	return s
}

func (s *Server) handleInitialize(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling initialize request from client: %v", request.Params)

	if initParams, ok := request.Params.AdditionalProperties.(mcp.InitializeRequestParams); !ok {
		return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Invalid initialize request parameters", nil)
	} else {
		s.clientCapabilities = &initParams.Capabilities
		s.clientVersion = &initParams.ClientInfo

		if slices.Contains(shared.SupportedProtocolVersions, initParams.ProtocolVersion) {
			s.clientVersion.Version = initParams.ProtocolVersion
		} else {
			s.logger.Warn("Client requested unsupported protocol version, falling back to latest supported version: %s", initParams.ProtocolVersion)
			s.clientVersion.Version = shared.LatestProtocolVersion
		}

		return jsonrpc.Result{
			AdditionalProperties: mcp.InitializeResult{
				ProtocolVersion: "1.0",
				Capabilities:    s.capabilities,
				ServerInfo:      s.serverInfo,
			},
		}, nil
	}
}

func (s *Server) OnInitialized(handler jsonrpc.NotificationHandler) {
	old := s.onInitialized
	s.onInitialized = func(notification *jsonrpc.JSONRPCNotification) error {
		if err := old(notification); err != nil {
			return err
		}
		if err := handler(notification); err != nil {
			return err
		}
		return nil
	}
}

const maxListResults = 100

func (s *Server) HandleListTools(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling list tools request from client: %v", request.Params)
	toolList := make([]mcp.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		toolList = append(toolList, tool.Tool)
	}

	var cursor *string
	if listParams, ok := request.Params.AdditionalProperties.(mcp.ListToolsRequestParams); ok {
		cursor = listParams.Cursor
	}

	toolList, nextCursor, err := paginate(request.Id, toolList, cursor)
	if err != nil {
		return jsonrpc.Result{}, err
	}

	return jsonrpc.Result{
		AdditionalProperties: mcp.ListToolsResult{
			Tools:      toolList,
			NextCursor: nextCursor,
		},
	}, nil
}

func (s *Server) handleListPrompts(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling list prompts request from client: %v", request.Params)
	promptList := make([]mcp.Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		promptList = append(promptList, prompt.Prompt)
	}

	var cursor *string
	if listParams, ok := request.Params.AdditionalProperties.(mcp.ListPromptsRequestParams); ok {
		cursor = listParams.Cursor
	}

	promptList, nextCursor, err := paginate(request.Id, promptList, cursor)
	if err != nil {
		return jsonrpc.Result{}, err
	}

	return jsonrpc.Result{
		AdditionalProperties: mcp.ListPromptsResult{
			Prompts:    promptList,
			NextCursor: nextCursor,
		},
	}, nil
}

func (s *Server) handleListResources(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling list resources request from client: %v", request.Params)
	resourceList := make([]mcp.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resourceList = append(resourceList, resource.Resource)
	}

	var cursor *string
	if listParams, ok := request.Params.AdditionalProperties.(mcp.ListResourcesRequestParams); ok {
		cursor = listParams.Cursor
	}

	resourceList, nextCursor, err := paginate(request.Id, resourceList, cursor)
	if err != nil {
		return jsonrpc.Result{}, err
	}

	return jsonrpc.Result{
		AdditionalProperties: mcp.ListResourcesResult{
			Resources:  resourceList,
			NextCursor: nextCursor,
		},
	}, nil
}

func (s *Server) HandleCallTool(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling call tool request from client: %v", request.Params)

	if callParams, ok := request.Params.AdditionalProperties.(mcp.CallToolRequestParams); !ok {
		return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Invalid call tool request parameters", nil)
	} else {
		if tool, ok := s.tools[callParams.Name]; !ok {
			return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Tool not found", nil)
		} else {
			toolResult, err := tool.Handler(callParams)
			if err != nil {
				return jsonrpc.Result{}, err
			}
			return jsonrpc.Result{
				AdditionalProperties: toolResult,
			}, nil
		}
	}
}

func (s *Server) handleGetPrompt(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling get prompt request from client: %v", request.Params)

	if getParams, ok := request.Params.AdditionalProperties.(mcp.GetPromptRequestParams); !ok {
		return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Invalid get prompt request parameters", nil)
	} else {
		if prompt, ok := s.prompts[getParams.Name]; !ok {
			return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Prompt not found", nil)
		} else {
			promptResult := prompt.MessageProvider(getParams)
			return jsonrpc.Result{
				AdditionalProperties: promptResult,
			}, nil
		}
	}
}

func (s *Server) handleReadResource(ctx context.Context, request *jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
	s.logger.Info("Handling read resource request from client: %v", request.Params)

	if readParams, ok := request.Params.AdditionalProperties.(mcp.ReadResourceRequestParams); !ok {
		return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Invalid read resource request parameters", nil)
	} else {
		if resource, ok := s.resources[readParams.Uri]; !ok {
			return jsonrpc.Result{}, jsonrpc.NewJSONRPCErrorError(request.Id, jsonrpc.InvalidParams, "Resource not found", nil)
		} else {
			resourceResult := resource.ReadHandler(readParams)
			return jsonrpc.Result{
				AdditionalProperties: resourceResult,
			}, nil
		}
	}
}

// func (s *Server) handleListResourceTemplates(ctx context.Context, request jsonrpc.JSONRPCRequest, extra jsonrpc.RequestHandlerExtra) (jsonrpc.Result, error) {
// }

func (c *Server) AssertCapabilityForMethod(method jsonrpc.Method) error {
	switch method {
	case shared.SamplingCreateMessageMethod:
		if c.clientCapabilities.Sampling == nil {
			return fmt.Errorf("client does nto support sampling (required for %s)", method)
		}
	case shared.RootsListMethod:
		if c.clientCapabilities.Roots == nil {
			return fmt.Errorf("client does not support roots (required for %s)", method)
		}
	}

	return nil
}

func (c *Server) AssertNotificationCapability(method jsonrpc.Method) error {
	// switch method {
	// case shared.LoggingMessageNotificationMethod:
	// 	if c.clientCapabilities.Logging == nil {
	// 		return fmt.Errorf("client does not support logging (required for %s)", method)
	// 	}
	// case shared.ResourceUpdatedNotificationMethod, shared.ResourceListChangedNotificationMethod:
	// 	if c.clientCapabilities.Resources == nil {
	// 		return fmt.Errorf("client does not support resources (required for %s)", method)
	// 	}
	// case shared.ToolListChangedNotificationMethod:
	// 	if c.clientCapabilities.Tools == nil {
	// 		return fmt.Errorf("client does not support tools (required for %s)", method)
	// 	}
	// case shared.PromptListChangedNotificationMethod:
	// 	if c.clientCapabilities.Prompts == nil {
	// 		return fmt.Errorf("client does not support prompts (required for %s)", method)
	// 	}
	// }

	return nil
}

func (c *Server) AssertRequestHandlerCapability(method jsonrpc.Method) error {
	switch method {
	// case shared.SamplingCreateMessageMethod:
	// 	if c.capabilities.Sampling == nil {
	// 		return fmt.Errorf("server does not support sampling (required for %s)", method)
	// 	}
	case shared.LoggingSetLevelMethod:
		if c.capabilities.Logging == nil {
			return fmt.Errorf("server does not support logging (required for %s)", method)
		}
	case shared.GetPromptsMethod, shared.ListPromptsMethod:
		if c.capabilities.Prompts == nil {
			return fmt.Errorf("server does not support prompts (required for %s)", method)
		}
	case shared.ListResourcesMethod, shared.ReadResourcesMethod, shared.ListResourcesTemplatesMethod:
		if c.capabilities.Resources == nil {
			return fmt.Errorf("server does not support resources (required for %s)", method)
		}
	case shared.ToolsCallMethod, shared.ToolsListMethod:
		if c.capabilities.Tools == nil {
			return fmt.Errorf("server does not support tools (required for %s)", method)
		}
	}

	return nil
}

// AddTool registers a single tool. This tool can then be called by the client
func (s *Server) AddTool(
	name string,
	description string,
	inputSchema mcp.ToolInputSchema,
	handler func(mcp.CallToolRequestParams) (mcp.CallToolResult, error),
) error {
	if s.capabilities.Tools == nil {
		return errors.New("Server does not support tools capability. Enable it in ServerOptions.")
	}

	s.logger.Info("Registering tool %s", name)
	s.tools[name] = RegisteredTool{
		Tool: mcp.Tool{
			Name:        name,
			Description: description,
			InputSchema: inputSchema,
		},
		Handler: handler,
	}

	return nil
}

// AddTools registers multiple tools at once.
func (s *Server) AddTools(toolsToAdd []RegisteredTool) error {
	if s.capabilities.Tools == nil {
		return errors.New("Server does not support tools capability.")
	}

	s.logger.Info("Registering %d tools", len(toolsToAdd))
	for _, rt := range toolsToAdd {
		s.logger.Info("Registering tool %s", rt.Tool.Name)
		s.tools[rt.Tool.Name] = rt
	}

	return nil
}

// AddPrompt registers a single prompt. The prompt can then be retrieved by the client.
func (s *Server) AddPrompt(prompt mcp.Prompt, promptProvider func(mcp.GetPromptRequestParams) mcp.GetPromptResult) error {
	if s.capabilities.Prompts == nil {
		return errors.New("Server does not support prompts capability.")
	}

	s.logger.Info("Registering prompt %s", prompt.Name)
	s.prompts[prompt.Name] = RegisteredPrompt{
		Prompt:          prompt,
		MessageProvider: promptProvider,
	}

	return nil
}

// AddPrompts registers multiple prompts at once.
func (s *Server) AddPrompts(promptsToAdd []RegisteredPrompt) error {
	if s.capabilities.Prompts == nil {
		return errors.New("Server does not support prompts capability.")
	}

	s.logger.Info("Registering %d prompts", len(promptsToAdd))
	for _, rp := range promptsToAdd {
		s.logger.Info("Registering prompt %s", rp.Prompt.Name)
		s.prompts[rp.Prompt.Name] = rp
	}

	return nil
}

// AddResource registers a single resource.
func (s *Server) AddResource(
	uri string,
	name string,
	description string,
	mimeType string,
	readHandler func(mcp.ReadResourceRequestParams) mcp.ReadResourceResult,
) error {
	if s.capabilities.Resources == nil {
		return errors.New("Server does not support resources capability.")
	}

	s.logger.Info("Registering resource %s at %s", name, uri)
	s.resources[uri] = RegisteredResource{
		Resource: mcp.Resource{
			Uri:         uri,
			Name:        name,
			Description: description,
			MimeType:    mimeType,
		},
		ReadHandler: readHandler,
	}

	return nil
}

// AddResources registers multiple resources at once.
func (s *Server) AddResources(resourcesToAdd []RegisteredResource) error {
	if s.capabilities.Resources == nil {
		return errors.New("Server does not support resources capability.")
	}

	s.logger.Info("Registering %d resources", len(resourcesToAdd))
	for _, r := range resourcesToAdd {
		s.logger.Info("Registering resource %s at %s", r.Resource.Name, r.Resource.Uri)
		s.resources[r.Resource.Uri] = r
	}

	return nil
}

// Ping sends a ping request to the client to check connectivity.
func (s *Server) Ping() error {
	return s.SendRequest(s.ctx, shared.PingMethod, nil, nil, nil)
}

// CreateMessage creates a message using the server's sampling capability.
func (s *Server) CreateMessage(params mcp.CreateMessageRequestParams, options *mcp.RequestOptions) (*mcp.CreateMessageResult, error) {
	result := &mcp.CreateMessageResult{}

	err := s.SendRequest(
		s.ctx,
		shared.SamplingCreateMessageMethod,
		&jsonrpc.JSONRPCRequestParams{
			AdditionalProperties: params,
		},
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)

	return result, err
}

// ListRoots lists the available "roots" from the client's perspective (if supported).
func (s *Server) ListRoots(options *mcp.RequestOptions) (*mcp.ListRootsResult, error) {
	result := &mcp.ListRootsResult{}

	err := s.SendRequest(
		s.ctx,
		shared.RootsListMethod,
		nil,
		&jsonrpc.Result{
			AdditionalProperties: result,
		},
		options)

	return result, err
}

// SendLoggingMessage sends a logging message notification to the client.
func (s *Server) SendLoggingMessage(params mcp.LoggingMessageNotificationParams) error {
	return s.SendNotification(
		shared.LoggingMessageNotificationMethod,
		&jsonrpc.JSONRPCNotificationParams{
			AdditionalProperties: params,
		},
	)
}

// SendResourceUpdated sends a resource-updated notification to the client.
func (s *Server) SendResourceUpdated(params mcp.ResourceUpdatedNotificationParams) error {
	return s.SendNotification(
		shared.ResourceUpdatedNotificationMethod,
		&jsonrpc.JSONRPCNotificationParams{
			AdditionalProperties: params,
		},
	)
}

// SendResourceListChanged sends a notification to the client indicating that the list of resources has changed.
func (s *Server) SendResourceListChanged() error {
	return s.SendNotification(shared.ResourceListChangedNotificationMethod, nil)
}

// SendToolListChanged sends a notification to the client indicating that the list of tools has changed.
func (s *Server) SendToolListChanged() error {
	return s.SendNotification(shared.ToolListChangedNotificationMethod, nil)
}

// SendPromptListChanged sends a notification to the client indicating that the list of prompts has changed.
func (s *Server) SendPromptListChanged() error {
	return s.SendNotification(shared.NotificationsPromptListChangedMethod, nil)
}

// RegisteredTool represents a registered tool on the server.
type RegisteredTool struct {
	Tool    mcp.Tool
	Handler func(mcp.CallToolRequestParams) (mcp.CallToolResult, error)
}

// RegisteredPrompt represents a registered prompt on the server.
type RegisteredPrompt struct {
	Prompt          mcp.Prompt
	MessageProvider func(mcp.GetPromptRequestParams) mcp.GetPromptResult
}

// RegisteredResource represents a registered resource on the server.
type RegisteredResource struct {
	Resource    mcp.Resource
	ReadHandler func(mcp.ReadResourceRequestParams) mcp.ReadResourceResult
}

// RegisteredResourceTemplate represents a registered resource template on the server.
type RegisteredResourceTemplate struct {
	Template mcp.ResourceTemplate
	Handler  func(mcp.ReadResourceRequestParams) mcp.ReadResourceResult
}

func paginate[T any](requestId jsonrpc.RequestId, items []T, cursor *string) ([]T, *string, *jsonrpc.JSONRPCErrorError) {
	start := 0
	end := len(items)

	if cursor != nil {
		cursor, err := strconv.Atoi(*cursor)
		if err != nil {
			return nil, nil, jsonrpc.NewJSONRPCErrorError(requestId, jsonrpc.InvalidParams, "Invalid cursor", nil)
		}
		start = cursor
	}

	// if there are more items than maxListResults, we set nextCursor to the next index
	// eg: if start = 0 & len(toolList) = 1000, we return toolList[0:100] and nextCursor = "100"
	if end > start+maxListResults {
		end = start + maxListResults
		nextCursorValue := strconv.Itoa(start + maxListResults)
		cursor = &nextCursorValue
	}

	return items[start:end], cursor, nil
}
