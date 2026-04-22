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

// DB_CreateApproval inserts a new approval record for a PO, Transfer, or Reject.
func DB_CreateApproval(a dto.ApprovalModel) error {
	if a.ID.IsZero() {
		a.ID = primitive.NewObjectID()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := dbConfigs.ApprovalCollection.InsertOne(context.Background(), a)
	return err
}

// DB_GetApprovalByRef returns the approval record for a given referenceType + referenceId.
func DB_GetApprovalByRef(refType, refId string) (*dto.ApprovalModel, error) {
	var approval dto.ApprovalModel
	filter := bson.M{"referenceType": refType, "referenceId": refId}
	err := dbConfigs.ApprovalCollection.FindOne(context.Background(), filter).Decode(&approval)
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// DB_GetApprovalByID fetches approval by its MongoDB ObjectID.
func DB_GetApprovalByID(id primitive.ObjectID) (*dto.ApprovalModel, error) {
	var approval dto.ApprovalModel
	err := dbConfigs.ApprovalCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&approval)
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// DB_UpdateApprovalStatus changes the approval status to APPROVED or REJECTED.
func DB_UpdateApprovalStatus(id primitive.ObjectID, status, approvedBy, notes string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"approvedBy": approvedBy,
			"approvedAt": time.Now(),
			"notes":      notes,
		},
	}
	_, err := dbConfigs.ApprovalCollection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

// DB_SearchApprovals returns paginated approvals filtered by query fields.
func DB_SearchApprovals(query dto.SearchApprovalQuery) ([]dto.ApprovalModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.ReferenceType != "" {
		filter["referenceType"] = query.ReferenceType
	}
	if query.ReferenceId != "" {
		filter["referenceId"] = query.ReferenceId
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}

	total, err := dbConfigs.ApprovalCollection.CountDocuments(ctx, filter)
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

	cursor, err := dbConfigs.ApprovalCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var approvals []dto.ApprovalModel
	if err = cursor.All(ctx, &approvals); err != nil {
		return nil, 0, err
	}
	return approvals, total, nil
}

// DB_IsApproved checks whether a given reference has an APPROVED approval record.
func DB_IsApproved(refType, refId string) (bool, error) {
	approval, err := DB_GetApprovalByRef(refType, refId)
	if err != nil {
		return false, fmt.Errorf("no approval record found for %s %s", refType, refId)
	}
	return approval.Status == dto.ApprovalApproved, nil
}
