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
var MedicineBatchCollection *mongo.Collection
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

	db := MongoClient.Database("SkinFirts")
	fmt.Println(db.Name())

	FeaturedLawyerCollection = db.Collection("Doctors")
	DoctorInfoCollection = db.Collection("doctor_info")
	AppointmentCollection = db.Collection("appointments")
	DoctorScheduleCollection = db.Collection("doctor_schedules")
	MedicineCollection = db.Collection("medicines")
	MedicineBatchCollection = db.Collection("medicine_batches")
	IdCounters = db.Collection("id_counters")
	MedicineOrderCollection = db.Collection("medicine_orders")
	FocusCollection = db.Collection("focus_categories")
	DoctorWeeklyScheduleCollection = db.Collection("doctor_weekly_schedules")
	DoctorAvailabilityCollection = db.Collection("doctor_availabilities")
	BillCollection = db.Collection("bills")
	HospitalBillCollection = db.Collection("hospital_bills")

	// Role-based user collections
	PatientCollection = db.Collection("patients")
	DoctorUserCollection = db.Collection("doctor_users")
	AdminUserCollection = db.Collection("admin_users")
	StaffUserCollection = db.Collection("staff_users")
	ReportCollection = db.Collection("reports")
	NotificationCollection = db.Collection("notifications")
	ServiceCollection = db.Collection("services")

	// New pharmacy modules
	BranchCollection = db.Collection("branches")
	SupplierCollection = db.Collection("suppliers")
	PurchaseOrderCollection = db.Collection("purchase_orders")
	GRNCollection = db.Collection("grn")
	StockTransferCollection = db.Collection("stock_transfers")

	// ERP/WMS audit + workflow collections
	StockMovementCollection = db.Collection("stock_movements")
	RejectStockCollection = db.Collection("reject_stock")
	SupplierBillCollection = db.Collection("supplier_bills")
	ApprovalCollection = db.Collection("approvals")

	return client
}
