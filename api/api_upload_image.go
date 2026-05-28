package api

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ImageUploadHandler handles image file uploads.
// In offline mode, images are stored on the local filesystem under ./uploads/images/
type ImageUploadHandler struct{}

func NewImageUploadHandler(_ ...interface{}) *ImageUploadHandler {
	return &ImageUploadHandler{}
}

// UploadImage handles POST /upload/image
// Saves the image to local disk and returns a publicly accessible URL.
func (h *ImageUploadHandler) UploadImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to retrieve image from form payload: " + err.Error(),
		})
	}

	// Create upload directory if it doesn't exist
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory: " + err.Error(),
		})
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	localPath := filepath.Join(uploadDir, filename)

	if err := c.SaveFile(fileHeader, localPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save image file: " + err.Error(),
		})
	}

	// Build the public URL
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "http://localhost:3000"
	}
	publicURL := fmt.Sprintf("%s/uploads/images/%s", serverHost, filename)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Image uploaded successfully",
		"image_url": publicURL,
	})
}
