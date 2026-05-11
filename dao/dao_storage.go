package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==========================================
// RACK DAOs
// ==========================================

func DB_CreateRack(rack dto.Rack) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rack.ID = primitive.NewObjectID()
	id, err := GenerateId(ctx, "racks", "RCK")
	if err != nil {
		return err
	}
	rack.RackId = id
	rack.CreatedAt = time.Now()
	rack.UpdatedAt = time.Now()

	_, err = dbConfigs.RackCollection.InsertOne(ctx, rack)
	return err
}

func DB_GetRacks(search string, activeOnly bool, page, limit int) ([]dto.Rack, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if activeOnly {
		filter["isActive"] = true
	}
	if search != "" {
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	total, err := dbConfigs.RackCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := dbConfigs.RackCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var racks []dto.Rack
	if err = cursor.All(ctx, &racks); err != nil {
		return nil, 0, err
	}

	return racks, total, nil
}

func DB_GetRackByID(rackId string) (*dto.Rack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var rack dto.Rack
	err := dbConfigs.RackCollection.FindOne(ctx, bson.M{"rackId": rackId}).Decode(&rack)
	if err != nil {
		return nil, err
	}
	return &rack, nil
}

func DB_UpdateRack(rack dto.Rack) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":        rack.Name,
			"description": rack.Description,
			"updatedAt":   time.Now(),
		},
	}
	res, err := dbConfigs.RackCollection.UpdateOne(ctx, bson.M{"rackId": rack.RackId}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("rack not found")
	}
	return nil
}

func DB_DeactivateRack(rackId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check for active shelves
	count, err := dbConfigs.ShelfCollection.CountDocuments(ctx, bson.M{"rackId": rackId, "isActive": true})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot deactivate rack with active shelves")
	}

	res, err := dbConfigs.RackCollection.UpdateOne(ctx, bson.M{"rackId": rackId}, bson.M{"$set": bson.M{"isActive": false, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("rack not found")
	}
	return nil
}

func DB_ActivateRack(rackId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := dbConfigs.RackCollection.UpdateOne(ctx, bson.M{"rackId": rackId}, bson.M{"$set": bson.M{"isActive": true, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("rack not found")
	}
	return nil
}

// ==========================================
// SHELF DAOs
// ==========================================

func DB_CreateShelf(shelf dto.Shelf) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify rack exists
	rackCount, err := dbConfigs.RackCollection.CountDocuments(ctx, bson.M{"rackId": shelf.RackId})
	if err != nil {
		return err
	}
	if rackCount == 0 {
		return errors.New("referenced rack does not exist")
	}

	shelf.ID = primitive.NewObjectID()
	id, err := GenerateId(ctx, "shelves", "SHF")
	if err != nil {
		return err
	}
	shelf.ShelfId = id
	shelf.CreatedAt = time.Now()
	shelf.UpdatedAt = time.Now()

	_, err = dbConfigs.ShelfCollection.InsertOne(ctx, shelf)
	return err
}

func DB_GetShelves(rackId string, activeOnly bool, page, limit int) ([]dto.Shelf, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if rackId != "" {
		filter["rackId"] = rackId
	}
	if activeOnly {
		filter["isActive"] = true
	}

	total, err := dbConfigs.ShelfCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := dbConfigs.ShelfCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var shelves []dto.Shelf
	if err = cursor.All(ctx, &shelves); err != nil {
		return nil, 0, err
	}

	return shelves, total, nil
}

func DB_GetShelfByID(shelfId string) (*dto.Shelf, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shelf dto.Shelf
	err := dbConfigs.ShelfCollection.FindOne(ctx, bson.M{"shelfId": shelfId}).Decode(&shelf)
	if err != nil {
		return nil, err
	}
	return &shelf, nil
}

func DB_UpdateShelf(shelf dto.Shelf) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":        shelf.Name,
			"description": shelf.Description,
			"updatedAt":   time.Now(),
		},
	}
	res, err := dbConfigs.ShelfCollection.UpdateOne(ctx, bson.M{"shelfId": shelf.ShelfId}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("shelf not found")
	}
	return nil
}

func DB_DeactivateShelf(shelfId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check for active locations
	count, err := dbConfigs.LocationCollection.CountDocuments(ctx, bson.M{"shelfId": shelfId, "isActive": true})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot deactivate shelf with active locations")
	}

	res, err := dbConfigs.ShelfCollection.UpdateOne(ctx, bson.M{"shelfId": shelfId}, bson.M{"$set": bson.M{"isActive": false, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("shelf not found")
	}
	return nil
}

func DB_ActivateShelf(shelfId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := dbConfigs.ShelfCollection.UpdateOne(ctx, bson.M{"shelfId": shelfId}, bson.M{"$set": bson.M{"isActive": true, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("shelf not found")
	}
	return nil
}

// ==========================================
// LOCATION DAOs
// ==========================================

func DB_CreateLocation(location dto.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check unique position index rule programmatically before DB
	count, err := dbConfigs.LocationCollection.CountDocuments(ctx, bson.M{"shelfId": location.ShelfId, "position": location.Position})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("position already exists in this shelf")
	}

	// Fetch Rack and Shelf to generate code
	var rack dto.Rack
	if err := dbConfigs.RackCollection.FindOne(ctx, bson.M{"rackId": location.RackId}).Decode(&rack); err != nil {
		return errors.New("invalid rackId")
	}

	var shelf dto.Shelf
	if err := dbConfigs.ShelfCollection.FindOne(ctx, bson.M{"shelfId": location.ShelfId}).Decode(&shelf); err != nil {
		return errors.New("invalid shelfId")
	}

	// Verify shelf belongs to rack
	if shelf.RackId != rack.RackId {
		return errors.New("shelf does not belong to the specified rack")
	}

	location.Code = fmt.Sprintf("%s%s-%02d", rack.Name, shelf.Name, location.Position)

	location.ID = primitive.NewObjectID()
	id, err := GenerateId(ctx, "locations", "LOC")
	if err != nil {
		return err
	}
	location.LocationId = id
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()

	_, err = dbConfigs.LocationCollection.InsertOne(ctx, location)
	return err
}

func DB_GetLocations(rackId, shelfId, searchCode string, activeOnly bool, page, limit int) ([]dto.Location, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if rackId != "" {
		filter["rackId"] = rackId
	}
	if shelfId != "" {
		filter["shelfId"] = shelfId
	}
	if activeOnly {
		filter["isActive"] = true
	}
	if searchCode != "" {
		filter["code"] = bson.M{"$regex": searchCode, "$options": "i"}
	}

	total, err := dbConfigs.LocationCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := dbConfigs.LocationCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var locations []dto.Location
	if err = cursor.All(ctx, &locations); err != nil {
		return nil, 0, err
	}

	return locations, total, nil
}

func DB_GetLocationByID(locationId string) (*dto.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var location dto.Location
	err := dbConfigs.LocationCollection.FindOne(ctx, bson.M{"locationId": locationId}).Decode(&location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func DB_GetLocationByCode(code string) (*dto.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var location dto.Location
	err := dbConfigs.LocationCollection.FindOne(ctx, bson.M{"code": code}).Decode(&location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func DB_UpdateLocation(location dto.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// We specifically DO NOT update Code, RackId, ShelfId, or Position 
	// to enforce immutability of coordinates after creation.
	update := bson.M{
		"$set": bson.M{
			"description": location.Description,
			"isOccupied":  location.IsOccupied,
			"updatedAt":   time.Now(),
		},
	}
	res, err := dbConfigs.LocationCollection.UpdateOne(ctx, bson.M{"locationId": location.LocationId}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("location not found")
	}
	return nil
}

func DB_DeactivateLocation(locationId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check for active batches
	count, err := dbConfigs.MedicineBatchCollection.CountDocuments(ctx, bson.M{"locationId": locationId, "status": "ACTIVE"})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot deactivate location linked to active batches")
	}

	res, err := dbConfigs.LocationCollection.UpdateOne(ctx, bson.M{"locationId": locationId}, bson.M{"$set": bson.M{"isActive": false, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("location not found")
	}
	return nil
}

func DB_ActivateLocation(locationId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := dbConfigs.LocationCollection.UpdateOne(ctx, bson.M{"locationId": locationId}, bson.M{"$set": bson.M{"isActive": true, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("location not found")
	}
	return nil
}
