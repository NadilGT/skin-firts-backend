package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_SavePatientFCMToken upserts the fcmToken field for a patient identified by their Firebase UID.
func DB_SavePatientFCMToken(firebaseUID string, fcmToken string) error {
	filter := bson.M{"firebaseUid": firebaseUID}
	update := bson.M{
		"$set": bson.M{
			"fcmToken": fcmToken,
		},
	}
	opts := options.Update().SetUpsert(false) // patient must already exist
	_, err := dbConfigs.PatientCollection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

// DB_GetPatientFCMToken fetches the FCM token of a patient by their Firebase UID.
// Returns an empty string if no token is stored (no error).
func DB_GetPatientFCMToken(firebaseUID string) (string, error) {
	var result struct {
		FcmToken string `bson:"fcmToken"`
	}

	filter := bson.M{"firebaseUid": firebaseUID}
	projection := bson.M{"fcmToken": 1, "_id": 0}
	opts := options.FindOne().SetProjection(projection)

	err := dbConfigs.PatientCollection.FindOne(context.Background(), filter, opts).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.FcmToken, nil
}
