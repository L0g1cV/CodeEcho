package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/L0g1cV/CodeEcho/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	storage *storage.MinIOClient
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(storage *storage.MinIOClient) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// UploadResponse is the response for file upload
type UploadResponse struct {
	URL string `json:"url"`
}

// AllowedImageExtensions are the allowed image file extensions
var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".ico":  true,
}

// MaxFileSize is the maximum file size (10MB)
const MaxFileSize = 10 << 20

// Upload handles image upload
// @Summary Upload an image
// @Description Upload an image file to object storage
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "Image file"
// @Success 200 {object} UploadResponse
// @Router /api/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 10MB)"})
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !AllowedImageExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to MinIO
	url, err := h.storage.UploadFile(c.Request.Context(), file, header.Size, contentType, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{URL: url})
}
