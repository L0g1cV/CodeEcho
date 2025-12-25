package handlers

import (
	"fmt"
	"net/http"

	"github.com/L0g1cV/CodeEcho/backend/internal/database"
	"github.com/L0g1cV/CodeEcho/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteRequest represents the request body for creating/updating a note
type NoteRequest struct {
	Type         string  `json:"type" binding:"required"` // "notebook", "folder", "note"
	Title        string  `json:"title" binding:"required"`
	ParentID     *string `json:"parentId"`     // null for notebooks, required for folders/notes
	Content      string  `json:"content"`      // Markdown content (for all types)
	ProblemURL   string  `json:"problemUrl"`   // Only for type="note"
	SolutionCode string  `json:"solutionCode"` // Only for type="note"
}

// NoteUpdateRequest represents the request body for updating a note
type NoteUpdateRequest struct {
	Title        *string `json:"title"`
	ParentID     *string `json:"parentId"`
	Content      *string `json:"content"`
	ProblemURL   *string `json:"problemUrl"`
	SolutionCode *string `json:"solutionCode"`
}

// NotesResponse represents notes response
type NotesResponse struct {
	Total int64         `json:"total"`
	Items []models.Note `json:"items"`
}

// GetNotes returns notes for the current user with optional filtering
func GetNotes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Query parameters
	noteType := c.Query("type")     // "notebook", "folder", "note", or empty for all
	parentID := c.Query("parentId") // Filter by parent
	search := c.Query("search")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "100")

	query := database.DB.Model(&models.Note{}).Where("user_id = ?", userID)

	// Filter by type
	if noteType != "" {
		query = query.Where("type = ?", noteType)
	}

	// Filter by parent
	if parentID != "" {
		if parentID == "null" || parentID == "root" {
			query = query.Where("parent_id IS NULL")
		} else {
			pID, err := uuid.Parse(parentID)
			if err == nil {
				query = query.Where("parent_id = ?", pID)
			}
		}
	}

	// Search in title and content
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Pagination
	var offset, limit int
	_, _ = parseIntQuery(page, &offset, 1)
	_, _ = parseIntQuery(pageSize, &limit, 100)
	offset = (offset - 1) * limit

	var notes []models.Note
	if err := query.Order("type ASC, updated_at DESC").Offset(offset).Limit(limit).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
		return
	}

	c.JSON(http.StatusOK, NotesResponse{
		Total: total,
		Items: notes,
	})
}

// GetNote returns a single note by ID
func GetNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// CreateNote creates a new note (can be notebook, folder, or note)
func CreateNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate type
	noteType := models.NoteType(req.Type)
	if noteType != models.NoteTypeNotebook && noteType != models.NoteTypeFolder && noteType != models.NoteTypeNote {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type. Must be 'notebook', 'folder', or 'note'"})
		return
	}

	// Notebooks must not have a parent (they are root level)
	if noteType == models.NoteTypeNotebook && req.ParentID != nil && *req.ParentID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notebooks cannot have a parent"})
		return
	}

	// Folders and notes must have a parent
	if noteType != models.NoteTypeNotebook && (req.ParentID == nil || *req.ParentID == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folders and notes must have a parent"})
		return
	}

	note := models.Note{
		UserID:       userID.(uuid.UUID),
		Type:         noteType,
		Title:        req.Title,
		Content:      req.Content,
		ProblemURL:   req.ProblemURL,
		SolutionCode: req.SolutionCode,
	}

	// Set parent if provided
	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent ID"})
			return
		}

		// Verify parent exists and belongs to user
		var parent models.Note
		if err := database.DB.Where("id = ? AND user_id = ?", parentID, userID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent not found"})
			return
		}

		// Validate hierarchy: notes can only have notebook/folder as parent, folders can only have notebook/folder as parent
		if parent.Type == models.NoteTypeNote {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create items inside a note"})
			return
		}

		note.ParentID = &parentID
	}

	if err := database.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// UpdateNote updates an existing note
func UpdateNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	var req NoteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	if req.Title != nil {
		note.Title = *req.Title
	}
	if req.Content != nil {
		note.Content = *req.Content
	}
	if req.ProblemURL != nil {
		note.ProblemURL = *req.ProblemURL
	}
	if req.SolutionCode != nil {
		note.SolutionCode = *req.SolutionCode
	}

	// Handle parent change (for moving items between folders)
	if req.ParentID != nil {
		// Notebooks cannot be moved (they have no parent)
		if note.Type == models.NoteTypeNotebook {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change parent of a notebook"})
			return
		}

		if *req.ParentID == "" {
			// Moving to root is not allowed for folders/notes
			c.JSON(http.StatusBadRequest, gin.H{"error": "Folders and notes must have a parent"})
			return
		}

		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent ID"})
			return
		}

		// Verify parent exists
		var parent models.Note
		if err := database.DB.Where("id = ? AND user_id = ?", parentID, userID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent not found"})
			return
		}

		// Prevent circular reference
		if parentID == id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot set self as parent"})
			return
		}

		// Cannot move into a note
		if parent.Type == models.NoteTypeNote {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot move items inside a note"})
			return
		}

		note.ParentID = &parentID
	}

	if err := database.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note (and recursively deletes children for notebooks/folders)
func DeleteNote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// For notebooks and folders, recursively delete all children
	if note.Type == models.NoteTypeNotebook || note.Type == models.NoteTypeFolder {
		if err := deleteChildrenRecursive(id, userID.(uuid.UUID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete children"})
			return
		}
	}

	// Delete the note itself
	if err := database.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.Status(http.StatusNoContent)
}

// deleteChildrenRecursive deletes all children of a note (for notebooks/folders)
func deleteChildrenRecursive(parentID uuid.UUID, userID uuid.UUID) error {
	var children []models.Note
	if err := database.DB.Where("parent_id = ? AND user_id = ?", parentID, userID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		// Recursively delete children of folders
		if child.Type == models.NoteTypeFolder {
			if err := deleteChildrenRecursive(child.ID, userID); err != nil {
				return err
			}
		}
		// Delete the child
		if err := database.DB.Delete(&child).Error; err != nil {
			return err
		}
	}

	return nil
}

// Helper function to parse int query parameters
func parseIntQuery(value string, target *int, defaultVal int) (bool, error) {
	if value == "" {
		*target = defaultVal
		return false, nil
	}
	var v int
	_, err := fmt.Sscanf(value, "%d", &v)
	if err != nil {
		*target = defaultVal
		return false, err
	}
	*target = v
	return true, nil
}
