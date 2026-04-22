package apiHandlers

import (
	"lawyerSL-Backend/api"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authMiddleware *AuthMiddleware, firebaseApp *firebase.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber on Render!")
	})

	// ========== APP DOWNLOAD ROUTE ==========
	app.Get("/download/app", func(c *fiber.Ctx) error {
		// Forces the browser to download the APK instead of returning it as text
		return c.Download("./uploads/app/skin_first_app.apk", "Skin_First.apk")
	})

	// ========== ROLE MANAGEMENT ROUTES ==========
	roleHandler := NewRoleAssignmentHandler(firebaseApp)
	staffHandler := api.NewStaffHandler(firebaseApp)
	imageUploadHandler := api.NewImageUploadHandler(firebaseApp)
	appointmentStatusHandler := api.NewAppointmentStatusHandler(firebaseApp)
	reportHandler := api.NewReportHandler(firebaseApp)

	// Admin-only role management routes
	app.Post("/admin/create-staff", authMiddleware.ValidateToken, RequiresRole("admin"), staffHandler.CreateStaffAccount)
	app.Get("/admin/search-staff", staffHandler.SearchStaff)
	app.Get("/admin/search-patients", api.SearchPatients)

	// Admin-only role management routes
	app.Post("/admin/assign-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.AssignRoles)
	app.Get("/admin/user-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.GetUserRoles)
	app.Get("/admin/list-users", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.ListAllUsers)
	app.Delete("/admin/remove-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.RemoveRoles)

	// ========== GLOBAL ASSET ROUTES ==========
	app.Post("/upload/image", authMiddleware.ValidateToken, imageUploadHandler.UploadImage)

	// ========== USER REGISTRATION ROUTES ==========
	// Patient registers themselves after Firebase sign-up (public — no token needed here,
	// but the FirebaseUID in the body ties the record to their auth identity).
	app.Post("/register/patient", api.CreatePatientUser)
	// Only an existing admin can onboard a new doctor or another admin.
	app.Post("/register/doctor-user", api.CreateDoctorUserAccount)
	app.Post("/register/admin", api.CreateAdminUser)

	// ========== ROLE LOOKUP ROUTES ==========
	// Portal: checks admin_users collection only — returns 404 if user is not an admin.
	app.Get("/role/admin", api.FindAdminRole)
	// Mobile app: checks patients + doctor_users collections.
	app.Get("/role/mobile", api.FindMobileUserRole)

	// ========== AUTH PROFILE ==========
	// Returns uid, email, branchId, roles from the JWT — useful for frontend after login.
	app.Get("/auth/me", authMiddleware.ValidateToken, api.GetMyProfile)

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
	app.Patch("/appointments/id/running", api.SetAppointmentRunning)
	app.Post("/create/appointment",authMiddleware.ValidateToken, BranchMiddleware,RequiresRole("admin"), api.CreateAppointment)
	// Branch-scoped: admins see their branch; super_admin sees all
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
	app.Get("/medicines/:id", api.GetMedicineByID)
	app.Put("/medicines/:id", api.UpdateMedicine)
	app.Delete("/medicines/:id", api.DeleteMedicine)

	// ========== MEDICINE BATCH ROUTES ==========
	app.Post("/batches", api.CreateMedicineBatch)
	app.Get("/batches/medicineId", api.GetBatchesByMedicineID)
	app.Get("/batches/available/medicineId", api.GetAvailableBatchesFEFO)
	app.Get("/batches/active-stock/medicineId", api.GetActiveStockByMedicineID)

	// ========== BILLING ROUTES ==========
	app.Post("/billing/deduct", authMiddleware.ValidateToken, BranchMiddleware, api.DeductStockFEFO)
	app.Post("/billing/create-bill", authMiddleware.ValidateToken, BranchMiddleware, api.CreateBill)
	app.Post("/billing/confirm/billId", authMiddleware.ValidateToken, BranchMiddleware, api.ConfirmBill)
	app.Get("/billing/pdf", api.GenerateBillPDF)

	// Hospital Bill Routes
	app.Post("/billing/hospital-bill", authMiddleware.ValidateToken, BranchMiddleware, api.CreateHospitalBill)
	app.Put("/billing/hospital-bill/confirm/:id", authMiddleware.ValidateToken, BranchMiddleware, api.ConfirmHospitalBill)
	app.Get("/billing/hospital-bill/:id/pdf", api.DownloadHospitalBillPDF)

	// ========== MEDICINE ORDER ROUTES ==========
	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)

	// ========== NEW DOCTOR SCHEDULING ROUTES ==========
	// Doctor Weekly Schedule
	app.Post("/doctor-weekly-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.CreateDoctorWeeklySchedule)
	app.Put("/doctor-weekly-schedule/doctorId", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateDoctorWeeklySchedule)
	app.Delete("/doctor-weekly-schedule/doctorId", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteDoctorWeeklySchedule)
	app.Get("/doctor-weekly-schedule", authMiddleware.ValidateToken, BranchMiddleware, api.GetAllDoctorWeeklySchedules)
	app.Get("/doctor-weekly-schedule/available-dates", api.GetDoctorAvailableDatesForWeek) // Publicly accessible

	// Doctor Availability
	app.Post("/doctor-availability", authMiddleware.ValidateToken, BranchMiddleware, api.CreateDoctorAvailability)
	app.Put("/doctor-availability/doctorAvailabilityId", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateDoctorAvailability)
	app.Delete("/doctor-availability/doctorAvailabilityId", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteDoctorAvailability)
	app.Get("/doctor-availability", authMiddleware.ValidateToken, BranchMiddleware, api.GetAllDoctorAvailabilities)
	app.Get("/doctor-availability/check", api.CheckDoctorAvailability) // Publicly accessible

	// ========== REPORT ROUTES ==========
	app.Post("/api/reports/upload", reportHandler.UploadReport)
	app.Get("/api/reports", reportHandler.GetReportsByPatientID)

	// ========== NOTIFICATION ROUTES ==========
	// Notifications are created INTERNALLY by the backend — not via a public endpoint.
	// Use functions.SaveAndSendNotification(...) wherever you trigger a notification.
	//
	// GET    /api/notifications?userId=&lastId=&limit= → cursor-based pagination (mobile)
	// PATCH  /api/notifications/:id/read              → mark single as read (mobile)
	// PATCH  /api/notifications/read-all?userId=      → mark all as read (mobile)
	app.Get("/api/notifications", authMiddleware.ValidateToken, api.GetNotifications)
	app.Patch("/api/notifications/:notificationId/read", authMiddleware.ValidateToken, api.MarkNotificationRead)
	app.Patch("/api/notifications/read-all", authMiddleware.ValidateToken, api.MarkAllNotificationsRead)

	// ========== BRANCH MANAGEMENT ==========
	app.Post("/admin/branches", authMiddleware.ValidateToken, api.CreateBranch)
	app.Get("/branches", api.GetAllBranches)
	app.Get("/branches/:id", api.GetBranchByID)
	app.Put("/branches/:id", authMiddleware.ValidateToken, api.UpdateBranch)
	app.Delete("/branches/:id", authMiddleware.ValidateToken, api.DeleteBranch)

	// ========== SUPPLIER MANAGEMENT (branch-scoped) ==========
	app.Post("/suppliers", authMiddleware.ValidateToken, BranchMiddleware, api.CreateSupplier)
	app.Get("/suppliers", authMiddleware.ValidateToken, BranchMiddleware, api.GetSuppliers)
	app.Get("/suppliers/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierByID)
	app.Put("/suppliers/:id", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateSupplier)
	app.Delete("/suppliers/:id", authMiddleware.ValidateToken, BranchMiddleware, api.DeleteSupplier)

	// ========== PURCHASE ORDERS (branch-scoped) ==========
	app.Post("/purchase-orders", authMiddleware.ValidateToken, BranchMiddleware, api.CreatePurchaseOrder)
	app.Get("/purchase-orders", authMiddleware.ValidateToken, BranchMiddleware, api.GetPurchaseOrders)
	app.Get("/purchase-orders/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetPurchaseOrderByID)
	app.Patch("/purchase-orders/:id/status", authMiddleware.ValidateToken, BranchMiddleware, api.UpdatePurchaseOrderStatus)

	// ========== GRN (branch-scoped) ==========
	app.Post("/grn", authMiddleware.ValidateToken, BranchMiddleware, api.CreateGRN)
	app.Get("/grn", authMiddleware.ValidateToken, BranchMiddleware, api.GetGRNs)
	app.Get("/grn/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetGRNByID)

	// ========== INVENTORY (branch-scoped) ==========
	app.Get("/inventory/stock-valuation", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockValuation)
	app.Get("/inventory/expiry-alerts", authMiddleware.ValidateToken, BranchMiddleware, api.GetExpiryAlerts)
	app.Get("/inventory/stock-report", authMiddleware.ValidateToken, BranchMiddleware, api.GetInventoryStockReport)

	// ========== STOCK TRANSFERS (branch-scoped) ==========
	app.Post("/stock-transfers", authMiddleware.ValidateToken, BranchMiddleware, api.CreateStockTransfer)
	app.Get("/stock-transfers", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockTransfers)
	app.Patch("/stock-transfers/:id/approve", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ApproveStockTransfer)
	app.Patch("/stock-transfers/:id/complete", authMiddleware.ValidateToken, BranchMiddleware, api.CompleteStockTransfer)
	app.Patch("/stock-transfers/:id/cancel", authMiddleware.ValidateToken, BranchMiddleware, api.CancelStockTransfer)

	// ========== PAYMENT MANAGEMENT (branch-scoped) ==========
	app.Get("/billing/pharmacy", authMiddleware.ValidateToken, BranchMiddleware, api.GetPharmacyBills)
	app.Get("/billing/pharmacy/:billId", authMiddleware.ValidateToken, BranchMiddleware, api.GetPharmacyBillByID)
	app.Patch("/billing/pharmacy/:billId/payment", authMiddleware.ValidateToken, BranchMiddleware, api.UpdatePharmacyBillPayment)
	app.Get("/payments/daily-summary", authMiddleware.ValidateToken, BranchMiddleware, api.GetDailySalesSummary)
	app.Get("/payments/revenue", authMiddleware.ValidateToken, BranchMiddleware, api.GetRevenueSummary)
	app.Get("/payments/pending", authMiddleware.ValidateToken, BranchMiddleware, api.GetPendingPayments)

	// ========== REPORTS & ANALYTICS (branch-scoped) ==========
	app.Get("/reports/top-selling", authMiddleware.ValidateToken, BranchMiddleware, api.GetTopSellingMedicines)
	app.Get("/reports/sales", authMiddleware.ValidateToken, BranchMiddleware, api.GetSalesReport)
	app.Get("/reports/profit-margin", authMiddleware.ValidateToken, BranchMiddleware, api.GetProfitMarginReport)
	app.Get("/reports/expiry", authMiddleware.ValidateToken, BranchMiddleware, api.GetExpiryReport)
	app.Get("/reports/stock", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockReportAnalytics)

	// ========== STOCK MOVEMENTS (audit ledger, read-only) ==========
	app.Get("/stock-movements", authMiddleware.ValidateToken, BranchMiddleware, api.GetStockMovements)
	app.Get("/stock-movements/batch/:batchId", authMiddleware.ValidateToken, BranchMiddleware, api.GetMovementsByBatch)

	// ========== REJECT STOCK (expired / damaged / returns) ==========
	app.Post("/reject-stock", authMiddleware.ValidateToken, BranchMiddleware, api.CreateRejectStock)
	app.Get("/reject-stock", authMiddleware.ValidateToken, BranchMiddleware, api.GetRejectStocks)
	app.Get("/reject-stock/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetRejectStockByID)
	app.Patch("/reject-stock/:id/approve", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ApproveRejectStock)
	app.Patch("/reject-stock/:id/execute", authMiddleware.ValidateToken, BranchMiddleware, RequiresRole("admin"), api.ExecuteRejectStock)

	// ========== SUPPLIER BILLS (invoices) ==========
	app.Post("/supplier-bills", authMiddleware.ValidateToken, BranchMiddleware, api.CreateSupplierBill)
	app.Get("/supplier-bills", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierBills)
	app.Get("/supplier-bills/:id", authMiddleware.ValidateToken, BranchMiddleware, api.GetSupplierBillByID)
	app.Patch("/supplier-bills/:id/payment", authMiddleware.ValidateToken, BranchMiddleware, api.UpdateSupplierBillPayment)

	// ========== APPROVALS (generic workflow) ==========
	app.Get("/approvals", authMiddleware.ValidateToken, BranchMiddleware, api.GetApprovals)
	app.Patch("/approvals/:id/approve", authMiddleware.ValidateToken, RequiresRole("admin"), api.ApproveRecord)
	app.Patch("/approvals/:id/reject", authMiddleware.ValidateToken, RequiresRole("admin"), api.RejectRecord)
}

