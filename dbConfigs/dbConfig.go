package dbConfigs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var FeaturedLawyerCollection *mongo.Collection
var DoctorInfoCollection *mongo.Collection
var AppointmentCollection *mongo.Collection
var DoctorScheduleCollection *mongo.Collection
var MedicineCollection *mongo.Collection
var MedicineBatchCollection *mongo.Collection // global batch identity (no qty/branch)
var BranchStockCollection *mongo.Collection   // branch-specific stock levels
var IdCounters *mongo.Collection
var MedicineOrderCollection *mongo.Collection
var FocusCollection *mongo.Collection
var DoctorWeeklyScheduleCollection *mongo.Collection
var DoctorAvailabilityCollection *mongo.Collection
var BillCollection *mongo.Collection
var HospitalBillCollection *mongo.Collection
var NotificationCollection *mongo.Collection

// Role-based user collections
var PatientCollection *mongo.Collection
var DoctorUserCollection *mongo.Collection
var AdminUserCollection *mongo.Collection
var StaffUserCollection *mongo.Collection
var ReportCollection *mongo.Collection
var ServiceCollection *mongo.Collection

// New pharmacy modules
var BranchCollection *mongo.Collection
var SupplierCollection *mongo.Collection
var PurchaseOrderCollection *mongo.Collection
var GRNCollection *mongo.Collection
var StockTransferCollection *mongo.Collection

// ERP/WMS audit + workflow collections
var StockMovementCollection *mongo.Collection
var RejectStockCollection *mongo.Collection
var SupplierBillCollection *mongo.Collection
var ApprovalCollection *mongo.Collection
var StockAdjustmentCollection *mongo.Collection

func ConnectMongoDB(uri string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB")
	MongoClient = client

	coreMedicalDb := client.Database("core_medical_db")
	userManagementDb := client.Database("user_management_db")
	pharmacyInventoryDb := client.Database("pharmacy_inventory_db")
	erpOperationsDb := client.Database("erp_operations_db")
	billingFinanceDb := client.Database("billing_finance_db")
	analyticsDb := client.Database("analytics_db")
	systemDb := client.Database("system_db")

	// core_medical_db
	FeaturedLawyerCollection = coreMedicalDb.Collection("Doctors")
	DoctorInfoCollection = coreMedicalDb.Collection("doctor_info")
	AppointmentCollection = coreMedicalDb.Collection("appointments")
	DoctorScheduleCollection = coreMedicalDb.Collection("doctor_schedules")
	DoctorWeeklyScheduleCollection = coreMedicalDb.Collection("doctor_weekly_schedules")
	DoctorAvailabilityCollection = coreMedicalDb.Collection("doctor_availabilities")
	FocusCollection = coreMedicalDb.Collection("focus_categories")
	ServiceCollection = coreMedicalDb.Collection("services")

	// pharmacy_inventory_db
	MedicineCollection = pharmacyInventoryDb.Collection("medicines")
	MedicineBatchCollection = pharmacyInventoryDb.Collection("medicine_batches")
	BranchStockCollection = pharmacyInventoryDb.Collection("branch_stock")
	StockMovementCollection = pharmacyInventoryDb.Collection("stock_movements")
	RejectStockCollection = pharmacyInventoryDb.Collection("reject_stock")
	StockAdjustmentCollection = pharmacyInventoryDb.Collection("stock_adjustments")

	// system_db
	IdCounters = systemDb.Collection("id_counters")

	// billing_finance_db
	BillCollection = billingFinanceDb.Collection("bills")
	HospitalBillCollection = billingFinanceDb.Collection("hospital_bills")
	MedicineOrderCollection = billingFinanceDb.Collection("medicine_orders")

	// user_management_db
	PatientCollection = userManagementDb.Collection("patients")
	DoctorUserCollection = userManagementDb.Collection("doctor_users")
	AdminUserCollection = userManagementDb.Collection("admin_users")
	StaffUserCollection = userManagementDb.Collection("staff_users")
	NotificationCollection = userManagementDb.Collection("notifications")

	// analytics_db
	ReportCollection = analyticsDb.Collection("reports")

	// erp_operations_db
	BranchCollection = erpOperationsDb.Collection("branches")
	SupplierCollection = erpOperationsDb.Collection("suppliers")
	PurchaseOrderCollection = erpOperationsDb.Collection("purchase_orders")
	GRNCollection = erpOperationsDb.Collection("grn")
	StockTransferCollection = erpOperationsDb.Collection("stock_transfers")
	SupplierBillCollection = erpOperationsDb.Collection("supplier_bills")
	ApprovalCollection = erpOperationsDb.Collection("approvals")

	return client
}
