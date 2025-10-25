package handlers

import (
	"net/http"

	"github.com/jamoowen/reminiscer/internal/api"
	"github.com/jamoowen/reminiscer/internal/errors"
	"github.com/jamoowen/reminiscer/internal/middleware"
	"github.com/jamoowen/reminiscer/internal/models"
	"github.com/labstack/echo/v4"
)

type GroupHandler struct {
	store   models.Store
	authMid *middleware.AuthMiddleware
}

func NewGroupHandler(store models.Store, authMid *middleware.AuthMiddleware) *GroupHandler {
	return &GroupHandler{
		store:   store,
		authMid: authMid,
	}
}

// Create handles creating a new group
func (h *GroupHandler) Create(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	// Verify all members exist
	for _, memberID := range req.Members {
		member, err := h.store.Users().GetByID(memberID)
		if err != nil {
			if errors.IsCode(err, errors.CodeNotFound) {
				return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Member not found")
			}
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to verify member")
		}
		if !member.Authenticated {
			return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Member not authenticated")
		}
	}

	// Add creator to members if not already included
	hasCreator := false
	for _, memberID := range req.Members {
		if memberID == user.ID {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		req.Members = append(req.Members, user.ID)
	}

	// Create group entries for each member
	var groups []*models.Group
	for _, memberID := range req.Members {
		group := &models.Group{
			GroupID:  req.GroupID,
			Name:     req.Name,
			MemberID: memberID,
		}
		if err := h.store.Groups().Create(group); err != nil {
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to create group")
		}
		groups = append(groups, group)
	}

	// Get username lookup function
	getUsernameFn := func(userID string) string {
		u, err := h.store.Users().GetByID(userID)
		if err != nil || u == nil {
			return "Unknown"
		}
		return u.Username
	}

	return api.SendSuccess(c, http.StatusCreated, toGroupResponse(groups, getUsernameFn))
}

// List handles retrieving all groups for the current user
func (h *GroupHandler) List(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	groups, err := h.store.Groups().GetByMemberID(user.ID)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendSuccess(c, http.StatusOK, []interface{}{}) // Empty list instead of error
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve groups")
	}

	// Group the groups by GroupID
	groupMap := make(map[string][]*models.Group)
	for _, g := range groups {
		groupMap[g.GroupID] = append(groupMap[g.GroupID], g)
	}

	// Convert to response format
	getUsernameFn := func(userID string) string {
		u, err := h.store.Users().GetByID(userID)
		if err != nil || u == nil {
			return "Unknown"
		}
		return u.Username
	}

	var responses []*GroupResponse
	for _, groupList := range groupMap {
		responses = append(responses, toGroupResponse(groupList, getUsernameFn))
	}

	return api.SendSuccess(c, http.StatusOK, responses)
}

// Update handles updating a group
func (h *GroupHandler) Update(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	groupID := c.Param("id")
	if groupID == "" {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Group ID is required")
	}

	var req UpdateGroupRequest
	if err := c.Bind(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Invalid request data")
	}

	// Check if user is a member of the group
	groups, err := h.store.Groups().GetByGroupID(groupID)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Group not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve group")
	}

	isMember := false
	for _, g := range groups {
		if g.MemberID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return api.SendError(c, http.StatusForbidden, errors.CodeForbidden, "Not authorized to update this group")
	}

	// Verify all members exist
	for _, memberID := range req.Members {
		member, err := h.store.Users().GetByID(memberID)
		if err != nil {
			if errors.IsCode(err, errors.CodeNotFound) {
				return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Member not found")
			}
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to verify member")
		}
		if !member.Authenticated {
			return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Member not authenticated")
		}
	}

	// Update each group entry
	for _, g := range groups {
		g.Name = req.Name
		if err := h.store.Groups().Update(g); err != nil {
			return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to update group")
		}
	}

	// Get updated group
	updatedGroups, err := h.store.Groups().GetByGroupID(groupID)
	if err != nil {
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve updated group")
	}

	getUsernameFn := func(userID string) string {
		u, err := h.store.Users().GetByID(userID)
		if err != nil || u == nil {
			return "Unknown"
		}
		return u.Username
	}

	return api.SendSuccess(c, http.StatusOK, toGroupResponse(updatedGroups, getUsernameFn))
}

// Delete handles deleting a group
func (h *GroupHandler) Delete(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return api.SendError(c, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required")
	}

	groupID := c.Param("id")
	if groupID == "" {
		return api.SendError(c, http.StatusBadRequest, errors.CodeInvalidInput, "Group ID is required")
	}

	// Check if user is a member of the group
	groups, err := h.store.Groups().GetByGroupID(groupID)
	if err != nil {
		if errors.IsCode(err, errors.CodeNotFound) {
			return api.SendError(c, http.StatusNotFound, errors.CodeNotFound, "Group not found")
		}
		return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to retrieve group")
	}

	isMember := false
	for _, g := range groups {
		if g.MemberID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return api.SendError(c, http.StatusForbidden, errors.CodeForbidden, "Not authorized to delete this group")
	}

	// Delete all group entries
	for _, g := range groups {
		if err := h.store.Groups().Delete(g.ID); err != nil {
			if !errors.IsCode(err, errors.CodeNotFound) {
				return api.SendError(c, http.StatusInternalServerError, errors.CodeDatabaseError, "Failed to delete group")
			}
		}
	}

	return api.SendSuccess(c, http.StatusOK, nil)
}

// SetupRoutes sets up the group routes
func (h *GroupHandler) SetupRoutes(e *echo.Echo) {
	groups := e.Group("/groups", h.authMid.Authenticate)
	groups.POST("", h.Create)
	groups.GET("", h.List)
	groups.PATCH("/:id", h.Update)
	groups.DELETE("/:id", h.Delete)
}
