package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ──────────────────────────────────────────────
//  Supplier CRUD
// ──────────────────────────────────────────────

func DB_CreateSupplier(supplier dto.SupplierModel) error {
	_, err := dbConfigs.SupplierCollection.InsertOne(context.Background(), supplier)
	return err
}

func DB_GetSupplierByID(id primitive.ObjectID) (*dto.SupplierModel, error) {
	var supplier dto.SupplierModel
	err := dbConfigs.SupplierCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&supplier)
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func DB_SearchSuppliers(query dto.SearchSupplierQuery) ([]dto.SupplierModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}
	if query.SearchTerm != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
			{"contactPerson": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
			{"phone": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
		}
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	total, err := dbConfigs.SupplierCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := dbConfigs.SupplierCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var suppliers []dto.SupplierModel
	if err = cursor.All(ctx, &suppliers); err != nil {
		return nil, 0, err
	}
	return suppliers, total, nil
}

func DB_UpdateSupplier(id primitive.ObjectID, supplier dto.SupplierModel) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"name":          supplier.Name,
			"contactPerson": supplier.ContactPerson,
			"phone":         supplier.Phone,
			"email":         supplier.Email,
			"address":       supplier.Address,
			"taxNo":         supplier.TaxNo,
			"paymentTerms":  supplier.PaymentTerms,
			"status":        supplier.Status,
			"updatedAt":     time.Now(),
		},
	}
	_, err := dbConfigs.SupplierCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_DeleteSupplier(id primitive.ObjectID) error {
	_, err := dbConfigs.SupplierCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// ──────────────────────────────────────────────
//  Purchase Order
// ──────────────────────────────────────────────

func DB_CreatePurchaseOrder(po dto.PurchaseOrderModel) error {
	_, err := dbConfigs.PurchaseOrderCollection.InsertOne(context.Background(), po)
	return err
}

func DB_GetPurchaseOrderByID(id primitive.ObjectID) (*dto.PurchaseOrderModel, error) {
	var po dto.PurchaseOrderModel
	err := dbConfigs.PurchaseOrderCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&po)
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func DB_SearchPurchaseOrders(query dto.SearchPOQuery) ([]dto.PurchaseOrderModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}
	if query.SupplierId != "" {
		filter["supplierId"] = query.SupplierId
	}
	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	total, err := dbConfigs.PurchaseOrderCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := dbConfigs.PurchaseOrderCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var pos []dto.PurchaseOrderModel
	if err = cursor.All(ctx, &pos); err != nil {
		return nil, 0, err
	}
	return pos, total, nil
}

func DB_UpdatePOStatus(id primitive.ObjectID, req dto.UpdatePOStatusRequest) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":    req.Status,
			"notes":     req.Notes,
			"updatedAt": time.Now(),
		},
	}
	_, err := dbConfigs.PurchaseOrderCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// ──────────────────────────────────────────────
//  GRN — Goods Received Note
// ──────────────────────────────────────────────

// DB_CreateGRN saves the GRN and auto-creates a medicine batch for each line item.
func DB_CreateGRN(grn dto.GRNModel) error {
	ctx := context.Background()

	// Insert the GRN document
	if _, err := dbConfigs.GRNCollection.InsertOne(ctx, grn); err != nil {
		return err
	}

	// Auto-create medicine batches for received items
	for _, item := range grn.Items {
		batchId, err := GenerateId(ctx, "medicine_batches", "BAT")
		if err != nil {
			return err
		}
		batch := dto.MedicineBatchModel{
			ID:              primitive.NewObjectID(),
			MedicineBatchId: batchId,
			MedicineID:      item.MedicineID,
			Quantity:        item.Quantity,
			ExpiryDate:      item.ExpiryDate,
			BuyingPrice:     item.BuyingPrice,
			SellingPrice:    item.SellingPrice,
			Status:          "ACTIVE",
			BatchNumber:     item.BatchNumber,
			SupplierId:      grn.SupplierId,
			BranchId:        grn.BranchId,
			ReceivedDate:    grn.ReceivedDate,
			Notes:           grn.Notes,
			CreatedAt:       time.Now(),
		}
		if _, err := dbConfigs.MedicineBatchCollection.InsertOne(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func DB_GetGRNByID(id primitive.ObjectID) (*dto.GRNModel, error) {
	var grn dto.GRNModel
	err := dbConfigs.GRNCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&grn)
	if err != nil {
		return nil, err
	}
	return &grn, nil
}

func DB_SearchGRNs(query dto.SearchGRNQuery) ([]dto.GRNModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}
	if query.SupplierId != "" {
		filter["supplierId"] = query.SupplierId
	}
	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.PoId != "" {
		filter["poId"] = query.PoId
	}
	total, err := dbConfigs.GRNCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := dbConfigs.GRNCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var grns []dto.GRNModel
	if err = cursor.All(ctx, &grns); err != nil {
		return nil, 0, err
	}
	return grns, total, nil
}
