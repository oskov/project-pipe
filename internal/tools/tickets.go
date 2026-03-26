package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
)

// ListTickets lists tickets in a project, optionally filtered by status.
type ListTickets struct {
	svc       service.TicketService
	projectID string
}

func NewListTickets(svc service.TicketService, projectID string) *ListTickets {
	return &ListTickets{svc: svc, projectID: projectID}
}

func (t *ListTickets) Name() string        { return "list_tickets" }
func (t *ListTickets) Description() string { return "List tickets in the project. Use this to check whether an incoming request relates to an existing ticket before creating a new one." }
func (t *ListTickets) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"description": "Filter by status: open, in_progress, done, closed. Omit to list all.",
				"enum": ["open", "in_progress", "done", "closed"]
			}
		},
		"required": []
	}`)
}

func (t *ListTickets) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	tickets, err := t.svc.List(ctx, t.projectID, store.TicketStatus(args.Status))
	if err != nil {
		return "", fmt.Errorf("list tickets: %w", err)
	}
	if len(tickets) == 0 {
		return "no tickets found in this project", nil
	}

	var sb strings.Builder
	for _, tk := range tickets {
		desc := tk.Description
		if len(desc) > 80 {
			desc = desc[:80] + "…"
		}
		fmt.Fprintf(&sb, "[%s] (%s) %s\n  %s\n", tk.ID, tk.Status, tk.Title, desc)
	}
	return sb.String(), nil
}

// CreateTicket creates a new ticket in the project.
type CreateTicket struct {
	svc       service.TicketService
	projectID string
}

func NewCreateTicket(svc service.TicketService, projectID string) *CreateTicket {
	return &CreateTicket{svc: svc, projectID: projectID}
}

func (t *CreateTicket) Name() string        { return "create_ticket" }
func (t *CreateTicket) Description() string { return "Create a new ticket in the project for the incoming request." }
func (t *CreateTicket) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"title":       {"type": "string", "description": "Short title for the ticket"},
			"description": {"type": "string", "description": "Full description: requirements, acceptance criteria, constraints"}
		},
		"required": ["title", "description"]
	}`)
}

func (t *CreateTicket) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	ticket, err := t.svc.Create(ctx, t.projectID, args.Title, args.Description)
	if err != nil {
		return "", fmt.Errorf("create ticket: %w", err)
	}
	return fmt.Sprintf("created ticket %s: %q", ticket.ID, ticket.Title), nil
}

// GetTicket retrieves full details of a single ticket.
type GetTicket struct {
	svc service.TicketService
}

func NewGetTicket(svc service.TicketService) *GetTicket {
	return &GetTicket{svc: svc}
}

func (t *GetTicket) Name() string        { return "get_ticket" }
func (t *GetTicket) Description() string { return "Get full details of a ticket by ID." }
func (t *GetTicket) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"ticket_id": {"type": "string", "description": "The ticket ID"}
		},
		"required": ["ticket_id"]
	}`)
}

func (t *GetTicket) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		TicketID string `json:"ticket_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	tk, err := t.svc.GetByID(ctx, args.TicketID)
	if err != nil {
		return "", fmt.Errorf("get ticket: %w", err)
	}
	return fmt.Sprintf("ID: %s\nTitle: %s\nStatus: %s\nDescription:\n%s",
		tk.ID, tk.Title, tk.Status, tk.Description), nil
}
