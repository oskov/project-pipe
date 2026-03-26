package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
)

type projectHandler struct {
	projects service.ProjectService
}

type createProjectRequest struct {
	Name       string `json:"name"`
	GithubRepo string `json:"github_repo"`
}

type projectResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	GithubRepo string    `json:"github_repo"`
	LocalPath  string    `json:"local_path"`
	CreatedAt  time.Time `json:"created_at"`
}

func toProjectResponse(p *store.Project) projectResponse {
	return projectResponse{
		ID:         p.ID,
		Name:       p.Name,
		GithubRepo: p.GithubRepo,
		LocalPath:  p.LocalPath,
		CreatedAt:  p.CreatedAt,
	}
}

func (h *projectHandler) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.projects.Create(r.Context(), req.Name, req.GithubRepo)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalid):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, toProjectResponse(p))
}

func (h *projectHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]projectResponse, len(projects))
	for i, p := range projects {
		resp[i] = toProjectResponse(p)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *projectHandler) getProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	p, err := h.projects.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

