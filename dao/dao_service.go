package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DB_CreateService inserts a new service record into the services collection.
func DB_CreateService(service *dto.ServiceModel) error {
	service.ID = primitive.NewObjectID()
	_, err := dbConfigs.ServiceCollection.InsertOne(context.Background(), service)
	return err
}

// DB_GetAllServices retrieves all service records from the services collection.
func DB_GetAllServices() ([]dto.ServiceModel, error) {
	var services []dto.ServiceModel

	cursor, err := dbConfigs.ServiceCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &services); err != nil {
		return nil, err
	}

	// Ensure empty slice instead of nil
	if services == nil {
		services = []dto.ServiceModel{}
	}

	return services, nil
}

// DB_UpdateService updates an existing service record by its serviceId string.
func DB_UpdateService(serviceId string, update dto.ServiceModel) error {
	filter := bson.M{"serviceId": serviceId}
	updateData := bson.M{
		"$set": bson.M{
			"name":        update.Name,
			"description": update.Description,
			"unitPrice":   update.UnitPrice,
		},
	}
	_, err := dbConfigs.ServiceCollection.UpdateOne(context.Background(), filter, updateData)
	return err
}

// DB_DeleteService removes a service record from the collection by its serviceId string.
func DB_DeleteService(serviceId string) error {
	_, err := dbConfigs.ServiceCollection.DeleteOne(context.Background(), bson.M{"serviceId": serviceId})
	return err
}

// DB_CheckServiceExists checks if a service with the given name or serviceId already exists.
func DB_CheckServiceExists(identifier string) (bool, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": identifier},
			{"serviceId": identifier},
		},
	}

	count, err := dbConfigs.ServiceCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DB_GetServiceByServiceId retrieves a service by its serviceId string.
func DB_GetServiceByServiceId(serviceId string) (*dto.ServiceModel, error) {
	var service dto.ServiceModel
	err := dbConfigs.ServiceCollection.FindOne(context.Background(), bson.M{"serviceId": serviceId}).Decode(&service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}
