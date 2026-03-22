package api

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
)

type ImageUploadHandler struct {
	App *firebase.App
}

func NewImageUploadHandler(app *firebase.App) *ImageUploadHandler {
	return &ImageUploadHandler{App: app}
}

func (h *ImageUploadHandler) UploadImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to retrieve image from form payload: " + err.Error(),
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded image file",
		})
	}
	defer file.Close()

	client, err := h.App.Storage(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize storage client",
		})
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to resolve default storage bucket. Check your FIREBASE_STORAGE_BUCKET env variable.",
		})
	}

	// Generate a unique filename using timestamp
	filename := fmt.Sprintf("profile_pics/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	
	obj := bucket.Object(filename)
	writer := obj.NewWriter(context.Background())
	
	// PredefinedACL sets the file to be publicly accessible directly via GCS URL
	writer.ObjectAttrs.PredefinedACL = "publicRead"

	// Pipe the file chunks right into the cloud!
	if _, err := io.Copy(writer, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to safely transmit chunks to the cloud bucket: " + err.Error(),
		})
	}
	
	if err := writer.Close(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to properly conclude streaming image constraints: " + err.Error(),
		})
	}

	// Grab the bucket metadata dynamically to construct the URL
	bucketAttrs, err := bucket.Attrs(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to extract bucket metadata bounds: " + err.Error(),
		})
	}

	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketAttrs.Name, filename)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Image transmitted successfully",
		"image_url": publicURL,
	})
}
