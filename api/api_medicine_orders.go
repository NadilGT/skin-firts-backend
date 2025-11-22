package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- New Medicine Order Handlers ---

// CreateMedicineOrder handles the creation of a new medicine order and stock deduction.
func CreateMedicineOrder(c *fiber.Ctx) error {
	var order dto.MedicineOrderModel
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// 1. Basic Validation
	if len(order.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order must contain at least one item"})
	}
	if order.PatientName == "" || order.ContactNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient name and contact number are required"})
	}

	// 2. Business Logic & Calculation
	totalAmount := 0.0
	for i := range order.Items {
		item := &order.Items[i]
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Item quantity must be greater than 0"})
		}
		
		// Ensure unit cost is present (should ideally be retrieved from MedicineModel during order entry on front end)
		// For simplicity, we trust the UnitCost provided here for now.
		item.TotalPrice = float64(item.Quantity) * item.UnitCost
		totalAmount += item.TotalPrice
	}
	order.TotalAmount = totalAmount

	// 3. Generate Receipt ID
	receiptID, err := dao.GenerateId(context.Background(), "medicineorders", "ORD")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate order ID: "+err.Error())
	}

	// 4. Set Metadata and Default Status
	order.ReceiptID = receiptID
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	if order.OrderStatus == "" {
		order.OrderStatus = "Pending" // Default status
	}

	// 5. Database Operation (Create Order & Update Stock)
	if err := dao.DB_CreateMedicineOrder(order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create medicine order: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Medicine order created successfully and stock updated.",
		"data": 	order,
	})
}

// GetMedicineOrder retrieves a single order by its MongoDB ObjectID
func GetMedicineOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order ID format"})
	}

	order, err := dao.DB_GetMedicineOrderByID(objectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Medicine order not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": order})
}

// SearchMedicineOrders retrieves a list of orders based on query parameters
func SearchMedicineOrders(c *fiber.Ctx) error {
	var query dto.SearchOrderQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	orders, total, err := dao.DB_SearchMedicineOrders(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search medicine orders"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": orders,
		"pagination": fiber.Map{
			"page": 		query.Page,
			"limit": 		query.Limit,
			"total": 		total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func UpdateMedicineOrderStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order ID format"})
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body for status update: " + err.Error(),
		})
	}
	
	if req.OrderStatus == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "New order status is required"})
	}

	if err := dao.DB_UpdateMedicineOrderStatus(objectID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order status: " + err.Error(),
		})
	}
	
	// Fetch the updated order to return to the client
	updatedOrder, err := dao.DB_GetMedicineOrderByID(objectID)
	if err != nil {
		// Log the error but still return success if the update was successful
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Order status updated successfully to " + req.OrderStatus,
			"note": "Could not fetch updated order data.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order status updated successfully to " + updatedOrder.OrderStatus,
		"data": updatedOrder,
	})
}