package llm

import (
	"context"
	"encoding/json"
)

// Role identifies the author of a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a single entry in a conversation.
type Message struct {
	Role    Role
	Content string

	// Set on RoleAssistant messages when the LLM requests tool calls.
	ToolCalls []ToolCall

	// Set on RoleTool messages (tool result).
	ToolCallID string
	ToolName   string
}

// ToolCall represents a single tool invocation requested by the LLM.
type ToolCall struct {
	ID   string
	Name string
	Args string // JSON-encoded arguments
}

// ToolDefinition describes a tool available to the LLM.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage // JSON Schema object
}

// Response is returned by ChatWithTools.
type Response struct {
	Content   string     // final text (empty when ToolCalls is set)
	ToolCalls []ToolCall // non-empty when the LLM wants to call tools
}

// Client is the abstraction over any LLM provider.
type Client interface {
	// Chat sends messages and returns the assistant's text reply.
	Chat(ctx context.Context, messages []Message) (string, error)

	// ChatWithTools sends messages with available tools and returns either a
	// text response or a list of tool calls to execute.
	ChatWithTools(ctx context.Context, messages []Message, tools []ToolDefinition) (Response, error)
}

