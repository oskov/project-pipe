package langchain

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oskov/project-pipe/internal/config"
	"github.com/oskov/project-pipe/internal/llm"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type client struct {
	model llms.Model
}

// New creates an LLM client based on the provider in config.
func New(cfg config.LLMConfig) (llm.Client, error) {
	var (
		model llms.Model
		err   error
	)

	switch cfg.Provider {
	case "openai":
		opts := []openai.Option{openai.WithModel(cfg.Model)}
		if cfg.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.APIKey))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
		}
		model, err = openai.New(opts...)
	case "anthropic":
		opts := []anthropic.Option{anthropic.WithModel(cfg.Model)}
		if cfg.APIKey != "" {
			opts = append(opts, anthropic.WithToken(cfg.APIKey))
		}
		model, err = anthropic.New(opts...)
	case "ollama":
		opts := []ollama.Option{ollama.WithModel(cfg.Model)}
		if cfg.BaseURL != "" {
			opts = append(opts, ollama.WithServerURL(cfg.BaseURL))
		}
		model, err = ollama.New(opts...)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q", cfg.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("init %s client: %w", cfg.Provider, err)
	}

	return &client{model: model}, nil
}

// Chat sends messages without tools and returns the text response.
func (c *client) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	resp, err := c.ChatWithTools(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ChatWithTools sends messages with optional tools and returns either text or
// tool call requests.
func (c *client) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDefinition) (llm.Response, error) {
	lcMessages, err := toLangChainMessages(messages)
	if err != nil {
		return llm.Response{}, err
	}

	opts := []llms.CallOption{}
	if len(tools) > 0 {
		opts = append(opts, llms.WithTools(toLangChainTools(tools)))
	}

	resp, err := c.model.GenerateContent(ctx, lcMessages, opts...)
	if err != nil {
		return llm.Response{}, fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return llm.Response{}, fmt.Errorf("empty response from LLM")
	}

	choice := resp.Choices[0]

	// Tool calls take priority over text content.
	if len(choice.ToolCalls) > 0 {
		calls := make([]llm.ToolCall, len(choice.ToolCalls))
		for i, tc := range choice.ToolCalls {
			calls[i] = llm.ToolCall{
				ID:   tc.ID,
				Name: tc.FunctionCall.Name,
				Args: tc.FunctionCall.Arguments,
			}
		}
		return llm.Response{ToolCalls: calls}, nil
	}

	return llm.Response{Content: choice.Content}, nil
}

// ── conversion helpers ──────────────────────────────────────────────────────

func toLangChainMessages(messages []llm.Message) ([]llms.MessageContent, error) {
	out := make([]llms.MessageContent, 0, len(messages))
	for _, m := range messages {
		mc, err := toLangChainMessage(m)
		if err != nil {
			return nil, err
		}
		out = append(out, mc)
	}
	return out, nil
}

func toLangChainMessage(m llm.Message) (llms.MessageContent, error) {
	switch m.Role {
	case llm.RoleSystem:
		return llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(m.Content)},
		}, nil

	case llm.RoleUser:
		return llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(m.Content)},
		}, nil

	case llm.RoleAssistant:
		parts := []llms.ContentPart{}
		if m.Content != "" {
			parts = append(parts, llms.TextPart(m.Content))
		}
		for _, tc := range m.ToolCalls {
			parts = append(parts, llms.ToolCall{
				ID:   tc.ID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Args,
				},
			})
		}
		return llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: parts}, nil

	case llm.RoleTool:
		return llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: m.ToolCallID,
					Name:       m.ToolName,
					Content:    m.Content,
				},
			},
		}, nil

	default:
		return llms.MessageContent{}, fmt.Errorf("unknown role: %q", m.Role)
	}
}

func toLangChainTools(tools []llm.ToolDefinition) []llms.Tool {
	out := make([]llms.Tool, len(tools))
	for i, t := range tools {
		var params any
		if len(t.Parameters) > 0 {
			_ = json.Unmarshal(t.Parameters, &params)
		}
		out[i] = llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		}
	}
	return out
}

