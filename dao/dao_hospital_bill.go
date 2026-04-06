package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DB_CreateHospitalBill inserts a new hospital bill record into the hospital_bills collection.
func DB_CreateHospitalBill(bill *dto.HospitalBillModel) error {
	if bill.ID == primitive.NilObjectID {
		bill.ID = primitive.NewObjectID()
	}
	_, err := dbConfigs.HospitalBillCollection.InsertOne(context.Background(), bill)
	return err
}

// DB_ConfirmHospitalBill updates a hospital bill's confirm status to true.
func DB_ConfirmHospitalBill(hospitalBillId string) error {
	filter := bson.M{"hospitalBillId": hospitalBillId}
	update := bson.M{"$set": bson.M{"confirm": true}}
	_, err := dbConfigs.HospitalBillCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// DB_GetHospitalBill gets a hospital bill by its bill ID.
func DB_GetHospitalBill(hospitalBillId string) (*dto.HospitalBillModel, error) {
	var bill dto.HospitalBillModel
	err := dbConfigs.HospitalBillCollection.FindOne(context.Background(), bson.M{"hospitalBillId": hospitalBillId}).Decode(&bill)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}
