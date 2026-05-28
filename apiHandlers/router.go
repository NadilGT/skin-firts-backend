package apiHandlers

import (
	"lawyerSL-Backend/api"
	localauth "lawyerSL-Backend/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes registers all application routes.
// Firebase dependency is fully removed — all auth is handled by local JWT.
func SetupRoutes(app *fiber.App, authMiddleware *AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber — Local Auth Mode 🔐")
	})

	// ========== APP DOWNLOAD ROUTE ==========
	app.Get("/download/app", func(c *fiber.Ctx) error {
		return c.Download("./uploads/app/skin_first_app.apk", "Skin_First.apk")
	})

	// ========== AUTH ROUTES (public) ==========
	app.Post("/auth/login", localauth.Login)
	app.Post("/auth/register", localauth.Register)

	// ========== AUTH ROUTES (protected) ==========
	app.Get("/auth/me", authMiddleware.ValidateToken, localauth.Me)
	app.Post("/auth/change-password", authMiddleware.ValidateToken, localauth.ChangePassword)

	// Admin-only auth management
	app.Post("/auth/set-password", authMiddleware.ValidateToken, RequiresRole("admin"), localauth.SetPassword)

	// ========== ROLE MANAGEMENT ROUTES (MongoDB-based) ==========
	roleHandler := NewRoleAssignmentHandler()

	app.Post("/admin/assign-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.AssignRoles)
	app.Get("/admin/user-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.GetUserRoles)
	app.Get("/admin/list-users", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.ListAllUsers)
	app.Delete("/admin/remove-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.RemoveRoles)
	app.Patch("/admin/user-status", authMiddleware.ValidateToken, RequiresRole("admin"), UpdateUserStatus)

	// ========== STAFF MANAGEMENT ==========
	staffHandler := api.NewStaffHandler()
	imageUploadHandler := api.NewImageUploadHandler()
	appointmentStatusHandler := api.NewAppointmentStatusHandler()
	reportHandler := api.NewReportHandler()

	app.Post("/admin/create-staff", authMiddleware.ValidateToken, RequiresRole("admin"), staffHandler.CreateStaffAccount)
	app.Get("/admin/search-staff", authMiddleware.ValidateToken, RequiresRole("admin"), staffHandler.SearchStaff)
	app.Get("/admin/search-patients", authMiddleware.ValidateToken, api.SearchPatients)

	// ========== GLOBAL ASSET ROUTES ==========
	app.Post("/upload/image", authMiddleware.ValidateToken, imageUploadHandler.UploadImage)

	// ========== USER REGISTRATION ROUTES ==========
	// These legacy endpoints remain for backward-compat.
	// New frontends should use POST /auth/register instead.
	app.Post("/register/patient", api.CreatePatientUser)
	app.Post("/register/doctor-user", api.CreateDoctorUserAccount)
	app.Post("/register/admin", api.CreateAdminUser)

	// ========== ROLE LOOKUP ROUTES ==========
	app.Get("/role/admin", api.FindAdminRole)
	app.Get("/role/mobile", api.FindMobileUserRole)

	// ========== FOCUS ROUTES ==========
	app.Post("/focus", authMiddleware.ValidateToken, api.CreateFocus)
	app.Get("/findAll/focus", api.GetAllFocuses)

	// ========== SERVICE ROUTES ==========
	app.Post("/admin/services", api.CreateService)
	app.Get("/services", api.GetAllServices)
	app.Put("/admin/services/serviceId", api.UpdateService)
	app.Delete("/admin/services/serviceId", api.DeleteService)

	// ========== DOCTOR ROUTES ==========
	app.Post("/doctor", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctor)
	app.Get("/doctors", authMiddleware.ValidateToken, api.FindAllDoctors)
	app.Get("/doctors/search", api.SearchDoctorInfo)
	app.Get("/findAll/doctors/focus", api.GetDoctorsByFocus)
	app.Get("/doctor-info", authMiddleware.ValidateToken, api.FindDoctorInfoByName)
	app.Get("/doctor-info/id", authMiddleware.ValidateToken, api.FindDoctorInfoByDoctorId)
	app.Put("/doctor-info/id", authMiddleware.ValidateToken, RequiresRole("admin"), api.UpdateDoctorInfoByDoctorId)
	app.Post("/doctor-info", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctorInfo)
	app.Patch("/doctor-info/id/assign-branch", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.AssignDoctorToBranch)
	app.Patch("/doctor-info/id/remove-branch", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.RemoveDoctorFromBranch)

	// Public doctor routes
	app.Get("/doctor-info/favorite", api.GetFavoriteDoctors)
	app.Put("/doctor-info/favorite", api.ToggleFavoriteDoctor)

	// ========== FCM TOKEN ROUTES ==========
	app.Post("/api/users/save-token", api.SaveFCMToken)

	// ========== APPOINTMENT ROUTES ==========
	app.Get("/appointment/next-number/doctorId", api.GetNextAppointmentNumber)
	app.Get("/appointments/running/doctorId", api.GetRunningAppointmentNumber)
	app.Patch("/appointments/id/running", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.SetAppointmentRunning)
	app.Post("/create/appointment", authMiddleware.ValidateToken, api.CreateAppointment)
	app.Get("/findAll/appointments", authMiddleware.ValidateToken, BranchMiddleware, api.GetAllAppointments)
	app.Get("/findAll/appointments/doctor", authMiddleware.ValidateToken, BranchMiddleware, api.GetAppointmentsByDoctorID)
	app.Get("/findAll/appointments/doctor/ordered", api.GetAppointmentsByDoctorIDSortedByNumber)
	app.Get("/findAll/appointments/doctor/detailed", api.GetAppointmentsByDoctorDateStatus)
	app.Get("/findAll/appointments/patient", api.GetAppointmentsByPatientID)
	app.Get("/appointments/id/appointmentId", api.GetAppointmentByID)
	app.Put("/appointments/id/reschedule", appointmentStatusHandler.RescheduleAppointment)
	app.Patch("/appointments/id/status", appointmentStatusHandler.UpdateAppointmentStatus)

	// ========== DOCTOR SCHEDULE ROUTES (Legacy) ==========
	app.Post("/doctor-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.CreateDoctorSchedule)
	app.Get("/doctor-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.GetDoctorSchedule)
	app.Get("/doctor-schedule/range", authMiddleware.ValidateToken, BranchMiddleware, api.GetDoctorScheduleByDateRange)
	app.Delete("/doctor-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteDoctorSchedule)
	app.Delete("/doctor-schedule/time-slot", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteTimeSlotFromSchedule)

	// ========== MEDICINE ROUTES ==========
	app.Post("/medicines", api.CreateMedicine)
	app.Get("/medicines/search", api.SearchMedicines)
	app.Get("/medicines/low-stock", api.GetLowStockMedicines)
	app.Get("/medicines/barcode", api.GetMedicineByBarcode)
	app.Get("/medicines/:medicineId/best-batch", api.GetBestBatchesForMedicine)
	app.Get("/medicines/:id", api.GetMedicineByID)
	app.Put("/medicines/:id", api.UpdateMedicine)
	app.Delete("/medicines/:id", api.DeleteMedicine)

	// ========== MEDICINE BATCH ROUTES ==========
	app.Post("/batches", api.CreateMedicineBatch)
	app.Put("/batches/:id", authMiddleware.ValidateToken, api.UpdateMedicineBatch)
	app.Get("/batches/medicineId", api.GetBatchesByMedicineID)
	app.Get("/batches/available/medicineId", api.GetAvailableBatchesFEFO)
	app.Get("/batches/active-stock/medicineId", api.GetActiveStockByMedicineID)

	// ========== BILLING ROUTES ==========
	app.Post("/billing/deduct", authMiddleware.ValidateToken, BranchMiddleware, api.DeductStockFEFO)
	app.Post("/billing/create-bill", authMiddleware.ValidateToken, BranchMiddleware, api.CreateBill)
	app.Post("/billing/cancel-bill", authMiddleware.ValidateToken, BranchMiddleware, api.CancelBill)
	app.Post("/billing/confirm/billId", authMiddleware.ValidateToken, BranchMiddleware, api.ConfirmBill)
	app.Get("/billing/pdf", api.GenerateBillPDF)

	// Hospital Bill Routes
	app.Post("/billing/hospital-bill", authMiddleware.ValidateToken, BranchMiddleware, api.CreateHospitalBill)
	app.Put("/billing/hospital-bill/confirm/id", authMiddleware.ValidateToken, BranchMiddleware, api.ConfirmHospitalBill)
	app.Get("/billing/hospital-bill/:id/pdf", api.DownloadHospitalBillPDF)

	// ========== MEDICINE ORDER ROUTES ==========
	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)

	// ========== NEW DOCTOR SCHEDULING ROUTES ==========
	app.Post("/doctor-weekly-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.CreateDoctorWeeklySchedule)
	app.Put("/doctor-weekly-schedule/doctorId", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateDoctorWeeklySchedule)
	app.Delete("/doctor-weekly-schedule/doctorId", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteDoctorWeeklySchedule)
	app.Get("/doctor-weekly-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.GetAllDoctorWeeklySchedules)
	app.Get("/doctor-weekly-schedule/available-dates", api.GetDoctorAvailableDatesForWeek)

	// Doctor Availability
	app.Get("/doctor-availability/check", api.CheckDoctorAvailability)

	// ========== DOCTOR DAILY CAPACITY ROUTES (admin-only) ==========
	app.Get("/doctor-daily-capacity", authMiddleware.ValidateToken, BranchMiddleware, api.GetAllDailyCapacities)
	app.Get("/doctor-daily-capacity/single", authMiddleware.ValidateToken, BranchMiddleware, api.GetSingleDailyCapacity)
	app.Get("/doctor-daily-capacity/by-id", authMiddleware.ValidateToken, BranchMiddleware, api.GetDailyCapacityByID)
	app.Post("/doctor-daily-capacity", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.CreateDailyCapacity)
	app.Put("/doctor-daily-capacity", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.UpdateDailyCapacity)
	app.Delete("/doctor-daily-capacity", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.DeleteDailyCapacity)

	// ========== REPORT ROUTES ==========
	app.Post("/api/reports/upload", reportHandler.UploadReport)
	app.Get("/api/reports", reportHandler.GetReportsByPatientID)

	// ========== NOTIFICATION ROUTES ==========
	app.Get("/api/notifications", authMiddleware.ValidateToken, api.GetNotifications)
	app.Patch("/api/notifications/:notificationId/read", authMiddleware.ValidateToken, api.MarkNotificationRead)
	app.Patch("/api/notifications/read-all", authMiddleware.ValidateToken, api.MarkAllNotificationsRead)

	// ========== BRANCH MANAGEMENT ==========
	app.Get("/api/branches/context", authMiddleware.ValidateToken, api.GetBranchContext)
	app.Post("/admin/branches", authMiddleware.ValidateToken, api.CreateBranch)
	app.Get("/branches", api.GetAllBranches)
	app.Get("/branches/id", api.GetBranchByID)
	app.Put("/branches/id", authMiddleware.ValidateToken, api.UpdateBranch)
	app.Delete("/branches/id", authMiddleware.ValidateToken, api.DeleteBranch)

	// ========== SUPPLIER MANAGEMENT (branch-scoped) ==========
	app.Post("/suppliers", authMiddleware.ValidateToken, BranchMiddleware, api.CreateSupplier)
	app.Get("/suppliers", authMiddleware.ValidateToken, BranchMiddleware, api.GetSuppliers)
	app.Get("/suppliers/id", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierByID)
	app.Put("/suppliers/id", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateSupplier)
	app.Delete("/suppliers/id", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteSupplier)

	// ========== SUPPLIER MEDICINE PRICE (branch-scoped) ==========
	app.Post("/supplier-medicine-price", authMiddleware.ValidateToken, BranchMiddleware, api.CreateSupplierMedicinePrice)
	app.Get("/supplier-medicine-price", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierMedicinePrices)
	app.Get("/supplier-medicine-price/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierMedicinePriceByID)
	app.Put("/supplier-medicine-price/:id", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateSupplierMedicinePrice)
	app.Delete("/supplier-medicine-price/:id", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteSupplierMedicinePrice)

	// ========== PURCHASE ORDERS (branch-scoped) ==========
	app.Post("/purchase-orders", authMiddleware.ValidateToken, BranchMiddleware, api.CreatePurchaseOrder)
	app.Get("/purchase-orders", authMiddleware.ValidateToken, BranchMiddleware, api.GetPurchaseOrders)
	app.Get("/purchase-orders/filter", authMiddleware.ValidateToken, BranchMiddleware, api.GetPurchaseOrdersByStatus)
	app.Get("/purchase-orders/id", authMiddleware.ValidateToken, BranchMiddleware, api.GetPurchaseOrderByID)
	app.Patch("/purchase-orders/id/status", authMiddleware.ValidateToken, BranchMiddleware, api.UpdatePurchaseOrderStatus)

	// ========== GRN (branch-scoped) ==========
	app.Post("/grn", authMiddleware.ValidateToken, BranchMiddleware, api.CreateGRN)
	app.Get("/grn", authMiddleware.ValidateToken, BranchMiddleware, api.GetGRNs)
	app.Get("/grn/id", authMiddleware.ValidateToken, BranchMiddleware, api.GetGRNByID)

	// ========== INVENTORY (branch-scoped) ==========
	app.Get("/inventory/stock-valuation", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockValuation)
	app.Get("/inventory/expiry-alerts", authMiddleware.ValidateToken, BranchMiddleware, api.GetExpiryAlerts)
	app.Get("/inventory/stock-report", authMiddleware.ValidateToken, BranchMiddleware, api.GetInventoryStockReport)
	// /inventory/stocks/enriched MUST be before /inventory/stocks
	app.Get("/inventory/stocks/enriched", authMiddleware.ValidateToken, BranchMiddleware, api.GetBranchStocksEnriched)
	app.Get("/inventory/stocks", authMiddleware.ValidateToken, BranchMiddleware, api.GetBranchStocks)

	// ========== STOCK TRANSFERS (branch-scoped) ==========
	app.Post("/stock-transfers", authMiddleware.ValidateToken, BranchMiddleware, api.CreateStockTransfer)
	app.Get("/stock-transfers", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockTransfers)
	app.Get("/stock-transfers/id", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockTransferByTransferID)
	app.Patch("/stock-transfers/id/approve", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ApproveStockTransfer)
	app.Patch("/stock-transfers/id/complete", authMiddleware.ValidateToken, BranchMiddleware, api.CompleteStockTransfer)
	app.Patch("/stock-transfers/id/cancel", authMiddleware.ValidateToken, BranchMiddleware, api.CancelStockTransfer)

	// ========== PAYMENT MANAGEMENT (branch-scoped) ==========
	app.Get("/billing/pharmacy", authMiddleware.ValidateToken, BranchMiddleware, api.GetPharmacyBills)
	app.Get("/billing/pharmacy/:billId", authMiddleware.ValidateToken, BranchMiddleware, api.GetPharmacyBillByID)
	app.Patch("/billing/pharmacy/:billId/payment", authMiddleware.ValidateToken, BranchMiddleware, api.UpdatePharmacyBillPayment)
	app.Get("/payments/daily-summary", authMiddleware.ValidateToken, BranchMiddleware, api.GetDailySalesSummary)
	app.Get("/payments/revenue", authMiddleware.ValidateToken, BranchMiddleware, api.GetRevenueSummary)
	app.Get("/payments/pending", authMiddleware.ValidateToken, BranchMiddleware, api.GetPendingPayments)
	app.Get("/payments/total-revenue", authMiddleware.ValidateToken, BranchMiddleware, api.GetTotalRevenue)

	// ========== REPORTS & ANALYTICS (branch-scoped) ==========
	app.Get("/reports/top-selling", authMiddleware.ValidateToken, BranchMiddleware, api.GetTopSellingMedicinesPDF)
	app.Get("/reports/sales", authMiddleware.ValidateToken, BranchMiddleware, api.GetSalesReport)
	app.Get("/reports/profit-margin", authMiddleware.ValidateToken, BranchMiddleware, api.GetProfitMarginReport)
	app.Get("/reports/expiry", authMiddleware.ValidateToken, BranchMiddleware, api.GetExpiryReport)
	app.Get("/reports/stock", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockReportAnalytics)

	app.Get("/billing/reports/bills", authMiddleware.ValidateToken, BranchMiddleware, api.GetBillsReport)
	app.Get("/billing/reports/hospital-bills", authMiddleware.ValidateToken, BranchMiddleware, api.GetHospitalBillsReport)

	// ========== DASHBOARD ANALYTICS (branch-scoped) ==========
	app.Get("/analytics/appointments", authMiddleware.ValidateToken, BranchMiddleware, api.GetAppointmentsAnalytics)
	app.Get("/analytics/revenue", authMiddleware.ValidateToken, BranchMiddleware, api.GetRevenueAnalytics)
	app.Get("/analytics/summary", authMiddleware.ValidateToken, BranchMiddleware, api.GetDashboardSummary)

	// ========== STOCK MOVEMENTS (audit ledger) ==========
	app.Get("/stock-movements", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockMovements)
	app.Get("/stock-movements/batch/:batchId", authMiddleware.ValidateToken, BranchMiddleware, api.GetMovementsByBatch)

	// ========== REJECT STOCK ==========
	app.Post("/reject-stock", authMiddleware.ValidateToken, BranchMiddleware, api.CreateRejectStock)
	app.Get("/reject-stock", authMiddleware.ValidateToken, BranchMiddleware, api.GetRejectStocks)
	app.Get("/reject-stock/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetRejectStockByID)
	app.Patch("/reject-stock/:id/approve", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ApproveRejectStock)
	app.Patch("/reject-stock/:id/execute", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ExecuteRejectStock)

	// ========== STOCK ADJUSTMENTS ==========
	app.Post("/stock-adjustments", authMiddleware.ValidateToken, BranchMiddleware, api.CreateStockAdjustment)
	app.Get("/stock-adjustments", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockAdjustments)
	app.Patch("/stock-adjustments/:id/approve", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ApproveStockAdjustment)
	app.Patch("/stock-adjustments/:id/execute", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ExecuteStockAdjustment)

	// ========== SUPPLIER BILLS ==========
	app.Post("/supplier-bills", authMiddleware.ValidateToken, BranchMiddleware, api.CreateSupplierBill)
	app.Get("/supplier-bills", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierBills)
	app.Get("/supplier-bills/id", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierBillByID)
	app.Patch("/supplier-bills/id/payment", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateSupplierBillPayment)

	// ========== APPROVALS ==========
	app.Get("/approvals", authMiddleware.ValidateToken, BranchMiddleware, api.GetApprovals)
	app.Patch("/approvals/:id/approve", authMiddleware.ValidateToken, RequiresRole("admin"), api.ApproveRecord)
	app.Patch("/approvals/:id/reject", authMiddleware.ValidateToken, RequiresRole("admin"), api.RejectRecord)

	// ========== STORAGE MANAGEMENT ==========
	app.Post("/api/racks", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateRack)
	app.Get("/api/racks", authMiddleware.ValidateToken, api.GetRacks)
	app.Get("/api/racks/:id", authMiddleware.ValidateToken, api.GetRackByID)
	app.Put("/api/racks/:id", authMiddleware.ValidateToken, RequiresRole("admin"), api.UpdateRack)
	app.Patch("/api/racks/:id/deactivate", authMiddleware.ValidateToken, RequiresRole("admin"), api.DeactivateRack)
	app.Patch("/api/racks/:id/activate", authMiddleware.ValidateToken, RequiresRole("admin"), api.ActivateRack)

	app.Post("/api/shelves", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateShelf)
	app.Get("/api/shelves", authMiddleware.ValidateToken, api.GetShelves)
	app.Get("/api/shelves/:id", authMiddleware.ValidateToken, api.GetShelfByID)
	app.Put("/api/shelves/:id", authMiddleware.ValidateToken, RequiresRole("admin"), api.UpdateShelf)
	app.Patch("/api/shelves/:id/deactivate", authMiddleware.ValidateToken, RequiresRole("admin"), api.DeactivateShelf)
	app.Patch("/api/shelves/:id/activate", authMiddleware.ValidateToken, RequiresRole("admin"), api.ActivateShelf)
	app.Get("/api/racks/:rackId/shelves", authMiddleware.ValidateToken, api.GetShelvesByRackID)

	app.Post("/api/locations", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateLocation)
	app.Get("/api/locations", authMiddleware.ValidateToken, api.GetLocations)
	app.Get("/api/locations/:id", authMiddleware.ValidateToken, api.GetLocationByID)
	app.Put("/api/locations/:id", authMiddleware.ValidateToken, RequiresRole("admin"), api.UpdateLocation)
	app.Patch("/api/locations/:id/deactivate", authMiddleware.ValidateToken, RequiresRole("admin"), api.DeactivateLocation)
	app.Patch("/api/locations/:id/activate", authMiddleware.ValidateToken, RequiresRole("admin"), api.ActivateLocation)
	app.Get("/api/shelves/:shelfId/locations", authMiddleware.ValidateToken, api.GetLocationsByShelfID)

	// MUST be before /api/locations/:id
	app.Get("/api/locations/:id/batches", authMiddleware.ValidateToken, api.GetBatchesByLocation)

	// Warehouse Map
	app.Get("/api/warehouse/map", authMiddleware.ValidateToken, api.GetWarehouseMap)
}
