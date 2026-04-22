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

// DB_CreateSupplierBill saves a new supplier invoice.
func DB_CreateSupplierBill(bill dto.SupplierBillModel) error {
	if bill.ID.IsZero() {
		bill.ID = primitive.NewObjectID()
	}
	_, err := dbConfigs.SupplierBillCollection.InsertOne(context.Background(), bill)
	return err
}

// DB_GetSupplierBillByID fetches a supplier bill by its string billId.
func DB_GetSupplierBillByID(billId string) (*dto.SupplierBillModel, error) {
	var bill dto.SupplierBillModel
	err := dbConfigs.SupplierBillCollection.FindOne(context.Background(), bson.M{"billId": billId}).Decode(&bill)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

// DB_SearchSupplierBills returns paginated supplier bills with optional filters.
func DB_SearchSupplierBills(query dto.SearchSupplierBillQuery) ([]dto.SupplierBillModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.SupplierId != "" {
		filter["supplierId"] = query.SupplierId
	}
	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.PurchaseOrderId != "" {
		filter["purchaseOrderId"] = query.PurchaseOrderId
	}
	if query.GrnId != "" {
		filter["grnId"] = query.GrnId
	}
	if query.PaymentStatus != "" {
		filter["paymentStatus"] = query.PaymentStatus
	}
	if query.From != "" || query.To != "" {
		dateFilter := bson.M{}
		if query.From != "" {
			if t, err := time.Parse(time.RFC3339, query.From); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if query.To != "" {
			if t, err := time.Parse(time.RFC3339, query.To); err == nil {
				dateFilter["$lte"] = t
			}
		}
		if len(dateFilter) > 0 {
			filter["createdAt"] = dateFilter
		}
	}

	total, err := dbConfigs.SupplierBillCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	findOpts := options.Find().
		SetSkip(int64((query.Page-1)*query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.SupplierBillCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var bills []dto.SupplierBillModel
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, 0, err
	}
	return bills, total, nil
}

// DB_UpdateSupplierBillPayment records a payment against a supplier bill and
// recalculates dueAmount and paymentStatus accordingly.
func DB_UpdateSupplierBillPayment(billId string, req dto.UpdateSupplierBillPaymentRequest) error {
	ctx := context.Background()

	// Fetch current bill
	bill, err := DB_GetSupplierBillByID(billId)
	if err != nil {
		return err
	}

	newPaid := bill.PaidAmount + req.PaidAmount
	if newPaid > bill.TotalAmount {
		newPaid = bill.TotalAmount
	}
	due := bill.TotalAmount - newPaid

	paymentStatus := "UNPAID"
	switch {
	case newPaid >= bill.TotalAmount:
		paymentStatus = "PAID"
	case newPaid > 0:
		paymentStatus = "PARTIAL"
	}

	updateFields := bson.M{
		"paidAmount":    newPaid,
		"dueAmount":     due,
		"paymentStatus": paymentStatus,
		"updatedAt":     time.Now(),
	}
	if req.PaymentMethod != "" {
		updateFields["paymentMethod"] = req.PaymentMethod
	}
	if req.Notes != "" {
		updateFields["notes"] = req.Notes
	}

	_, err = dbConfigs.SupplierBillCollection.UpdateOne(
		ctx,
		bson.M{"billId": billId},
		bson.M{"$set": updateFields},
	)
	return err
}
