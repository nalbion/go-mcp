package shared

import "github.com/nalbion/go-mcp/pkg/jsonrpc"

const (
	InitializeMethod                      jsonrpc.Method = "initialize"
	InitializedMethod                     jsonrpc.Method = "initialized"
	PingMethod                            jsonrpc.Method = "ping"
	ListResourcesMethod                   jsonrpc.Method = "resources/list"
	ListResourcesTemplatesMethod          jsonrpc.Method = "resources/templates/list"
	ReadResourcesMethod                   jsonrpc.Method = "resources/read"
	ResourcesSubscribeMethod              jsonrpc.Method = "resources/subscribe"
	ResourcesUnsubscribeMethod            jsonrpc.Method = "resources/unsubscribe"
	ListPromptsMethod                     jsonrpc.Method = "prompts/list"
	GetPromptsMethod                      jsonrpc.Method = "prompts/get"
	NotificationsCancelledMethod          jsonrpc.Method = "notifications/cancelled"
	NotificationsInitializedMethod        jsonrpc.Method = "notifications/initialized"
	NotificationsProgressMethod           jsonrpc.Method = "notifications/progress"
	LoggingMessageNotificationMethod      jsonrpc.Method = "notifications/message"
	ResourceUpdatedNotificationMethod     jsonrpc.Method = "notifications/resources/updated"
	ResourceListChangedNotificationMethod jsonrpc.Method = "notifications/resources/list_changed"
	ToolListChangedNotificationMethod     jsonrpc.Method = "notifications/tools/list_changed"
	NotificationsRootsListChangedMethod   jsonrpc.Method = "notifications/roots/list_changed"
	NotificationsPromptListChangedMethod  jsonrpc.Method = "notifications/prompts/list_changed"
	ToolsListMethod                       jsonrpc.Method = "tools/list"
	ToolsCallMethod                       jsonrpc.Method = "tools/call"
	LoggingSetLevelMethod                 jsonrpc.Method = "logging/setLevel"
	SamplingCreateMessageMethod           jsonrpc.Method = "sampling/createMessage"
	CompletionCompleteMethod              jsonrpc.Method = "completion/complete"
	RootsListMethod                       jsonrpc.Method = "roots/list"
)

const (
	LatestProtocolVersion = "2024-11-05"
)

var SupportedProtocolVersions = []string{
	LatestProtocolVersion,
	"2024-10-07",
}
