package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/skills"
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

const defaultMaxIterations = 15

// Agent is a generic LLM agent with a system prompt, tools, and a skills registry.
type Agent struct {
	agentType     store.AgentType
	systemPrompt  string
	llm           llm.Client
	agentRuns     store.AgentRunRepository
	tools         []tools.Tool
	skillRegistry *skills.Registry
	maxIterations int
	logger        *slog.Logger
}

// Option configures an Agent.
type Option func(*Agent)

func WithTools(tt ...tools.Tool) Option {
	return func(a *Agent) { a.tools = append(a.tools, tt...) }
}

// WithSkills attaches a skills registry to the agent. The agent will receive a
// list of available skills in its system prompt and can fetch them on demand
// via the built-in get_skill tool.
func WithSkills(registry *skills.Registry) Option {
	return func(a *Agent) { a.skillRegistry = registry }
}

func WithMaxIterations(n int) Option {
	return func(a *Agent) { a.maxIterations = n }
}

func WithLogger(l *slog.Logger) Option {
	return func(a *Agent) { a.logger = l }
}

// New creates a new Agent.
func New(
	agentType store.AgentType,
	systemPrompt string,
	llmClient llm.Client,
	agentRuns store.AgentRunRepository,
	opts ...Option,
) *Agent {
	a := &Agent{
		agentType:     agentType,
		systemPrompt:  systemPrompt,
		llm:           llmClient,
		agentRuns:     agentRuns,
		maxIterations: defaultMaxIterations,
		logger:        slog.Default(),
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Run executes the ReAct loop for the given task and optional project.
// projectID may be empty — in that case memory tools are not attached.
func (a *Agent) Run(ctx context.Context, taskID, projectID, userPrompt string) (string, error) {
	// Propagate taskID through context so delegation tools (RunAgent) can
	// pass it to child agent runs for tracing.
	ctx = tools.ContextWithTaskID(ctx, taskID)

	runID := uuid.New().String()

	run := &store.AgentRun{
		ID:        runID,
		TaskID:    taskID,
		AgentType: a.agentType,
		Status:    store.AgentRunStatusRunning,
		Input:     userPrompt,
		StartedAt: time.Now().UTC(),
	}
	if err := a.agentRuns.Create(ctx, run); err != nil {
		return "", fmt.Errorf("create agent run: %w", err)
	}

	result, err := a.react(ctx, runID, projectID, userPrompt)
	if err != nil {
		_ = a.agentRuns.Fail(ctx, runID, err.Error())
		return "", err
	}

	if err := a.agentRuns.Complete(ctx, runID, result); err != nil {
		return "", fmt.Errorf("complete agent run: %w", err)
	}
	return result, nil
}

// react is the inner ReAct loop.
func (a *Agent) react(ctx context.Context, runID, projectID, userPrompt string) (string, error) {
	systemPrompt := a.systemPrompt
	allTools := append([]tools.Tool(nil), a.tools...)

	if a.skillRegistry != nil && !a.skillRegistry.Empty() {
		systemPrompt += a.skillRegistry.List()
		allTools = append(allTools, tools.NewGetSkill(a.skillRegistry))
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: userPrompt},
	}

	toolDefs := tools.ToDefinitions(allTools)
	toolMap := make(map[string]tools.Tool, len(allTools))
	for _, t := range allTools {
		toolMap[t.Name()] = t
	}

	for i := range a.maxIterations {
		a.logger.Debug("agent iteration",
			"run_id", runID,
			"agent", a.agentType,
			"iteration", i+1,
			"messages", len(messages),
		)

		resp, err := a.llm.ChatWithTools(ctx, messages, toolDefs)
		if err != nil {
			return "", fmt.Errorf("llm call (iteration %d): %w", i+1, err)
		}

		// No tool calls → final answer.
		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		// Append the assistant's tool-call message.
		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each requested tool call and collect results.
		for _, tc := range resp.ToolCalls {
			result := a.executeTool(ctx, toolMap, tc)
			a.logger.Debug("tool result",
				"run_id", runID,
				"tool", tc.Name,
				"result_len", len(result),
			)
			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
				ToolName:   tc.Name,
			})
		}
	}

	return "", fmt.Errorf("agent reached max iterations (%d) without a final answer", a.maxIterations)
}

// executeTool runs a single tool call and always returns a string (errors are
// returned as a string so the LLM can reason about them).
func (a *Agent) executeTool(ctx context.Context, toolMap map[string]tools.Tool, tc llm.ToolCall) string {
	t, ok := toolMap[tc.Name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %q", tc.Name)
	}
	result, err := t.Execute(ctx, tc.Args)
	if err != nil {
		return fmt.Sprintf("error executing %s: %s", tc.Name, err.Error())
	}
	return result
}

