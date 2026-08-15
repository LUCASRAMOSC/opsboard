package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/domain"
	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/store"
)

type WorkspaceStore interface {
	CreateWorkspace(
		ctx context.Context,
		name string,
	) (domain.Workspace, error)

	GetWorkspace(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Workspace, error)

	ListWorkspaces(
		ctx context.Context,
	) ([]domain.Workspace, error)
}

type WorkspaceHandler struct {
	store WorkspaceStore
}

func NewWorkspaceHandler(store WorkspaceStore) *WorkspaceHandler {
	return &WorkspaceHandler{
		store: store,
	}
}

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

type workspaceResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	var request createWorkspaceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workspace name is required",
		})

		return
	}

	workspace, err := h.store.CreateWorkspace(
		c.Request.Context(),
		name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create workspace",
		})

		return
	}

	c.JSON(
		http.StatusCreated,
		newWorkspaceResponse(workspace),
	)
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid workspace ID",
		})

		return
	}

	workspace, err := h.store.GetWorkspace(
		c.Request.Context(),
		id,
	)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workspace not found",
		})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workspace",
		})

		return
	}

	c.JSON(
		http.StatusOK,
		newWorkspaceResponse(workspace),
	)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	workspaces, err := h.store.ListWorkspaces(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list workspaces",
		})

		return
	}

	response := make(
		[]workspaceResponse,
		0,
		len(workspaces),
	)

	for _, workspace := range workspaces {
		response = append(
			response,
			newWorkspaceResponse(workspace),
		)
	}

	c.JSON(http.StatusOK, response)
}

func newWorkspaceResponse(
	workspace domain.Workspace,
) workspaceResponse {
	return workspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt,
		UpdatedAt: workspace.UpdatedAt,
	}
}
