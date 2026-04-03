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
var NotificationCollection *mongo.Collection

// Role-based user collections
var PatientCollection *mongo.Collection
var DoctorUserCollection *mongo.Collection
var AdminUserCollection *mongo.Collection
var StaffUserCollection *mongo.Collection
var ReportCollection *mongo.Collection
var ServiceCollection *mongo.Collection

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

	// Role-based user collections
	PatientCollection = db.Collection("patients")
	DoctorUserCollection = db.Collection("doctor_users")
	AdminUserCollection = db.Collection("admin_users")
	StaffUserCollection = db.Collection("staff_users")
	ReportCollection = db.Collection("reports")
	NotificationCollection = db.Collection("notifications")
	ServiceCollection = db.Collection("services")

	return client
}
