package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oskov/project-pipe/internal/service"
)

type taskHandler struct {
	tasks service.TaskService
}

type createTaskRequest struct {
	ProjectID string `json:"project_id"`
	Prompt    string `json:"prompt"`
}

type createTaskResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type getTaskResponse struct {
	TaskID    string  `json:"task_id"`
	ProjectID string  `json:"project_id"`
	Status    string  `json:"status"`
	TicketID  *string `json:"ticket_id,omitempty"`
}

func (h *taskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.tasks.Create(r.Context(), req.ProjectID, req.Prompt)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalid):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, createTaskResponse{
		TaskID: task.ID,
		Status: string(task.Status),
	})
}

func (h *taskHandler) getTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.tasks.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, getTaskResponse{
		TaskID:    task.ID,
		ProjectID: task.ProjectID,
		Status:    string(task.Status),
		TicketID:  task.TicketID,
	})
}
