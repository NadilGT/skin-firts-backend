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

func ConnectMongoDB(uri string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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

	return client
}
