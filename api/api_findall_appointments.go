package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func GetAllAppointments(c *fiber.Ctx) error {
	appointments, err := dao.DB_FindAllAppointments()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch appointments",
		})
	}

	if len(appointments) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":      "No appointments found",
			"appointments": []dto.AppointmentModel{},
		})
	}
	return c.Status(fiber.StatusOK).JSON(appointments)
}

func GetAppointmentsByDoctorID(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	appointments, err := dao.DB_FindAppointmentsByDoctorID(doctorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch appointments for this doctor",
		})
	}

	if len(appointments) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":      "No appointments found",
			"appointments": []dto.AppointmentModel{},
		})
	}
	return c.Status(fiber.StatusOK).JSON(appointments)
}
