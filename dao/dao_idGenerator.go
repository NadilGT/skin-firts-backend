package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GenerateId(ctx context.Context, collectionName string, prefix string) (string, error) {
	counterCollection := dbConfigs.IdCounters
	filter := bson.M{"_id": collectionName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).
		SetReturnDocument(options.After)
	var result struct {
		Seq int `bson:"seq"`
	}
	err := counterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to increment and get new value: %v", err)
	}
	newID := fmt.Sprintf("%s-%03d", prefix, result.Seq)
	return newID, nil
}


