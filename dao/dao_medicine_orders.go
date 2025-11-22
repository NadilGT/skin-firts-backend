package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- New Medicine Order Functions ---

// DB_CreateMedicineOrder inserts the order and deducts stock for each item.
func DB_CreateMedicineOrder(order dto.MedicineOrderModel) error {
	ctx := context.Background()

	// 1. Create the Order document
	if _, err := dbConfigs.MedicineOrderCollection.InsertOne(ctx, order); err != nil {
		return errors.New("failed to insert medicine order: " + err.Error())
	}

	// 2. Deduct Stock for each item
	for _, item := range order.Items {
		
		// Find the medicine by its unique MedicineId (not the MongoDB ObjectID)
		filter := bson.M{"medicineid": item.MedicineID}
		
		// Atomically decrease the stock quantity
		update := bson.M{
			"$inc": bson.M{"stockQuantity": -item.Quantity},
			"$set": bson.M{"updatedAt": time.Now()},
		}

		// The update should ideally be done within a transaction to guarantee consistency
		// but using UpdateOne with $inc provides atomicity for a single field update.
		result, err := dbConfigs.MedicineCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			// In a real system, you would need to implement rollback logic here.
			return errors.New("failed to update stock for " + item.Name + ": " + err.Error())
		}
		
		if result.ModifiedCount == 0 {
			// This happens if the medicine ID was not found, or the stock was already too low
			// (though MongoDB allows negative $inc).
			return errors.New("failed to update stock quantity or medicine not found for ID: " + item.MedicineID)
		}
	}

	return nil
}

// DB_SearchMedicineOrders retrieves a list of orders based on query parameters
func DB_SearchMedicineOrders(query dto.SearchOrderQuery) ([]dto.MedicineOrderModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.ReceiptID != "" {
		filter["receiptid"] = query.ReceiptID
	}
	if query.PatientName != "" {
		filter["patientName"] = bson.M{"$regex": query.PatientName, "$options": "i"}
	}
	if query.ContactNumber != "" {
		filter["contactNumber"] = query.ContactNumber
	}
	if query.Status != "" {
		filter["orderStatus"] = query.Status
	}

	total, err := dbConfigs.MedicineOrderCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64((query.Page - 1) * query.Limit))
	findOptions.SetLimit(int64(query.Limit))
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Latest orders first

	cursor, err := dbConfigs.MedicineOrderCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []dto.MedicineOrderModel
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func DB_GetMedicineOrderByID(id primitive.ObjectID) (*dto.MedicineOrderModel, error) {
	var order dto.MedicineOrderModel
	err := dbConfigs.MedicineOrderCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func DB_UpdateMedicineOrderStatus(id primitive.ObjectID, req dto.UpdateOrderStatusRequest) error {
	filter := bson.M{"_id": id}
	
	updateFields := bson.M{
		"orderStatus": req.OrderStatus,
		"updatedAt":   time.Now(),
	}

	if req.DispensedByUserID != "" {
		updateFields["dispensedByUserId"] = req.DispensedByUserID
	}
	
	update := bson.M{
		"$set": updateFields,
	}

	result, err := dbConfigs.MedicineOrderCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	
	if result.ModifiedCount == 0 {
		return errors.New("order not found or status already set")
	}
	
	return nil
}