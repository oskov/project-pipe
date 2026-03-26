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

// taskPRsHandler handles PR-related endpoints scoped to a task.
type taskPRsHandler struct {
prs service.PullRequestService
}

type prResponse struct {
ID           string `json:"id"`
GithubNumber int    `json:"github_number"`
Title        string `json:"title"`
URL          string `json:"url"`
HeadBranch   string `json:"head_branch"`
BaseBranch   string `json:"base_branch"`
Status       string `json:"status"`
}

func (h *taskPRsHandler) listTaskPRs(w http.ResponseWriter, r *http.Request) {
taskID := chi.URLParam(r, "id")
prs, err := h.prs.GetByTaskID(r.Context(), taskID)
if err != nil {
writeError(w, http.StatusInternalServerError, err.Error())
return
}
resp := make([]prResponse, len(prs))
for i, pr := range prs {
resp[i] = prResponse{
ID:           pr.ID,
GithubNumber: pr.GithubNumber,
Title:        pr.Title,
URL:          pr.URL,
HeadBranch:   pr.HeadBranch,
BaseBranch:   pr.BaseBranch,
Status:       string(pr.Status),
}
}
writeJSON(w, http.StatusOK, resp)
}
