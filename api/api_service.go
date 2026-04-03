package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

// CreateService handles POST /admin/services
func CreateService(c *fiber.Ctx) error {
	var service dto.ServiceModel
	if err := c.BodyParser(&service); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if service.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	// Generate ServiceID
	id, err := dao.GenerateId(context.Background(), "services", "SRV")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate service ID",
		})
	}
	service.ServiceID = id

	if err := dao.DB_CreateService(&service); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create service",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Service created successfully",
		"service": service,
	})
}

// GetAllServices handles GET /services
func GetAllServices(c *fiber.Ctx) error {
	services, err := dao.DB_GetAllServices()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve services",
		})
	}

	return c.Status(fiber.StatusOK).JSON(services)
}

// UpdateService handles PUT /admin/services/:serviceId
func UpdateService(c *fiber.Ctx) error {
	serviceId := c.Query("serviceId")
	if serviceId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	var update dto.ServiceModel
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := dao.DB_UpdateService(serviceId, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update service",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Service updated successfully",
	})
}

// DeleteService handles DELETE /admin/services/:serviceId
func DeleteService(c *fiber.Ctx) error {
	serviceId := c.Query("serviceId")
	if serviceId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	if err := dao.DB_DeleteService(serviceId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete service",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Service deleted successfully",
	})
}
