package dao

import (
	"context"
	"fmt"
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

func DB_UpdatePOStatus(id primitive.ObjectID, req dto.UpdatePOStatusRequest, approvedBy string) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":    req.Status,
			"notes":     req.Notes,
			"updatedAt": time.Now(),
		},
	}
	_, err := dbConfigs.PurchaseOrderCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// When a PO is approved, create an Approval record so GRN can gate on it
	if req.Status == "APPROVED" {
		// Fetch the PO to get its business ID
		po, poErr := DB_GetPurchaseOrderByID(id)
		if poErr == nil && po != nil {
			approvalId, aErr := GenerateId(ctx, "approvals", "APR")
			if aErr == nil {
				_ = DB_CreateApproval(dto.ApprovalModel{
					ID:            primitive.NewObjectID(),
					ApprovalId:    approvalId,
					ReferenceType: dto.ApprovalRefPO,
					ReferenceId:   po.PoId,
					Status:        dto.ApprovalApproved,
					ApprovedBy:    approvedBy,
					ApprovedAt:    time.Now(),
					Notes:         req.Notes,
					CreatedAt:     time.Now(),
				})
			}
		}
	}
	return nil
}

// ──────────────────────────────────────────────
//  GRN — Goods Received Note
// ──────────────────────────────────────────────

// DB_GetPurchaseOrderByPoId fetches a PO by its string PoId
func DB_GetPurchaseOrderByPoId(poId string) (*dto.PurchaseOrderModel, error) {
	var po dto.PurchaseOrderModel
	err := dbConfigs.PurchaseOrderCollection.FindOne(context.Background(), bson.M{"poId": poId}).Decode(&po)
	if err != nil {
		return nil, err
	}
	return &po, nil
}

// DB_CreateGRN saves the GRN, auto-creates a medicine batch for each line item,
// and writes a PURCHASE StockMovement to the audit ledger.
// If the GRN references a PO (PoId is set), the PO must be APPROVED or PARTIALLY_RECEIVED.
func DB_CreateGRN(grn dto.GRNModel) error {
	ctx := context.Background()

	var po *dto.PurchaseOrderModel
	if grn.PoId != "" {
		// fetch PO to check status
		var err error
		po, err = DB_GetPurchaseOrderByPoId(grn.PoId)
		if err != nil {
			return fmt.Errorf("PO check failed: %v", err)
		}
		if po.Status != "APPROVED" && po.Status != "PARTIALLY_RECEIVED" {
			return fmt.Errorf("purchase order %s must be APPROVED or PARTIALLY_RECEIVED before creating a GRN", grn.PoId)
		}
	}

	// Insert the GRN document
	if _, err := dbConfigs.GRNCollection.InsertOne(ctx, grn); err != nil {
		return err
	}

	// Auto set PO to PARTIALLY_RECEIVED
	if po != nil && po.Status == "APPROVED" {
		_ = DB_UpdatePOStatus(po.ID, dto.UpdatePOStatusRequest{Status: "PARTIALLY_RECEIVED", Notes: "System updated from GRN creation"}, grn.ReceivedBy)
	}

	// Auto-create medicine batches (global) + branch stock (per-branch) and write PURCHASE movements
	for _, item := range grn.Items {
		batchId, err := GenerateId(ctx, "medicine_batches", "BAT")
		if err != nil {
			return err
		}
		// 1. Insert global MedicineBatch — no qty, no branch
		batch := dto.MedicineBatch{
			ID:           primitive.NewObjectID(),
			BatchId:      batchId,
			MedicineId:   item.MedicineID,
			BatchNumber:  item.BatchNumber,
			ExpiryDate:   item.ExpiryDate,
			BuyingPrice:  item.BuyingPrice,
			SellingPrice: item.SellingPrice,
			SupplierId:   grn.SupplierId,
			Status:       "ACTIVE",
			Notes:        grn.Notes,
			CreatedAt:    time.Now(),
		}
		if err := dao_CreateMedicineBatch(ctx, batch); err != nil {
			return err
		}

		// 2. Insert BranchStock — branch-specific qty
		stockId, err := GenerateId(ctx, "branch_stock", "STK")
		if err != nil {
			return err
		}
		stock := dto.BranchStock{
			ID:               primitive.NewObjectID(),
			StockId:          stockId,
			BatchId:          batchId,
			MedicineId:       item.MedicineID,
			BranchId:         grn.BranchId,
			Quantity:         item.Quantity,
			ReservedQuantity: 0,
			UpdatedAt:        time.Now(),
		}
		if _, err := dbConfigs.BranchStockCollection.InsertOne(ctx, stock); err != nil {
			return err
		}

		// 3. Write PURCHASE movement to audit ledger
		movementId, err := GenerateId(ctx, "stock_movements", "MOV")
		if err != nil {
			return fmt.Errorf("failed to generate movement id: %v", err)
		}
		movement := dto.StockMovementModel{
			ID:            primitive.NewObjectID(),
			MovementId:    movementId,
			BatchId:       batchId,
			MedicineId:    item.MedicineID,
			BranchId:      grn.BranchId,
			Type:          dto.MovementPurchase,
			Quantity:      item.Quantity,
			ReferenceId:   grn.GrnId,
			ReferenceType: "GRN",
			Notes:         fmt.Sprintf("GRN receipt — batch %s, supplier %s", item.BatchNumber, grn.SupplierId),
			CreatedBy:     grn.ReceivedBy,
			CreatedAt:     time.Now(),
		}
		if err := DB_CreateStockMovement(movement); err != nil {
			return fmt.Errorf("batch created but failed to write movement: %v", err)
		}
	}
	return nil
}

// dao_CreateMedicineBatch is an internal helper that inserts a MedicineBatch (used within DAO layer).
func dao_CreateMedicineBatch(ctx context.Context, batch dto.MedicineBatch) error {
	_, err := dbConfigs.MedicineBatchCollection.InsertOne(ctx, batch)
	return err
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
