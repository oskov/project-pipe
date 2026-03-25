package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oskov/project-pipe/internal/store"
)

// TicketService defines business operations for project tickets.
type TicketService interface {
	List(ctx context.Context, projectID string, status store.TicketStatus) ([]*store.Ticket, error)
	Create(ctx context.Context, projectID, title, description string) (*store.Ticket, error)
	GetByID(ctx context.Context, ticketID string) (*store.Ticket, error)
}

type ticketService struct {
	repo store.TicketRepository
}

func NewTicketService(repo store.TicketRepository) TicketService {
	return &ticketService{repo: repo}
}

func (s *ticketService) List(ctx context.Context, projectID string, status store.TicketStatus) ([]*store.Ticket, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalid)
	}
	tickets, err := s.repo.ListByProject(ctx, projectID, status)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return tickets, nil
}

func (s *ticketService) Create(ctx context.Context, projectID, title, description string) (*store.Ticket, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalid)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalid)
	}

	now := time.Now().UTC()
	ticket := &store.Ticket{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      store.TicketStatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return ticket, nil
}

func (s *ticketService) GetByID(ctx context.Context, ticketID string) (*store.Ticket, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("%w: ticket_id is required", ErrInvalid)
	}
	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	}
	return ticket, nil
}
