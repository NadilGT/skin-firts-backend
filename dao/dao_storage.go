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

func DB_GetRacks(branchId, search string, activeOnly bool, page, limit int) ([]dto.Rack, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if branchId != "" {
		filter["branchId"] = branchId
	}
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

func DB_GetShelves(branchId, rackId string, activeOnly bool, page, limit int) ([]dto.Shelf, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if branchId != "" {
		filter["branchId"] = branchId
	}
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

func DB_GetLocations(branchId, rackId, shelfId, searchCode string, activeOnly bool, page, limit int) ([]dto.Location, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if branchId != "" {
		filter["branchId"] = branchId
	}
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

// ==========================================
// WAREHOUSE MAP & DASHBOARD DAOs
// ==========================================

// DB_GetBatchesByLocation returns all medicine batches currently assigned to a
// specific physical location slot. Used for the "drill-into-cell" view on the rack map UI.
func DB_GetBatchesByLocation(locationId string) ([]dto.MedicineBatch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"locationId": locationId}
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []dto.MedicineBatch
	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

// DB_GetWarehouseMap builds and returns the full physical storage tree:
//   Rack → Shelf → Location → [BatchSummary with per-branch stock]
//
// It fetches all active racks, shelves, locations and batches, then
// assembles the tree in-memory. This keeps the logic simple and readable —
// pharmacy storage collections are small (typically < 1000 documents each).
func DB_GetWarehouseMap(branchId string) ([]dto.RackWithShelves, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Fetch ALL racks (active + inactive so admins can see the full map)
	rackFilter := bson.M{}
	if branchId != "" {
		rackFilter["branchId"] = branchId
	}
	rackCursor, err := dbConfigs.RackCollection.Find(ctx, rackFilter,
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("fetch racks: %w", err)
	}
	defer rackCursor.Close(ctx)
	var allRacks []dto.Rack
	if err = rackCursor.All(ctx, &allRacks); err != nil {
		return nil, err
	}

	// 2. Fetch ALL shelves
	shelfFilter := bson.M{}
	if branchId != "" {
		shelfFilter["branchId"] = branchId
	}
	shelfCursor, err := dbConfigs.ShelfCollection.Find(ctx, shelfFilter,
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("fetch shelves: %w", err)
	}
	defer shelfCursor.Close(ctx)
	var allShelves []dto.Shelf
	if err = shelfCursor.All(ctx, &allShelves); err != nil {
		return nil, err
	}

	// 3. Fetch ALL locations
	locFilter := bson.M{}
	if branchId != "" {
		locFilter["branchId"] = branchId
	}
	locationCursor, err := dbConfigs.LocationCollection.Find(ctx, locFilter,
		options.Find().SetSort(bson.D{{Key: "position", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("fetch locations: %w", err)
	}
	defer locationCursor.Close(ctx)
	var allLocations []dto.Location
	if err = locationCursor.All(ctx, &allLocations); err != nil {
		return nil, err
	}

	// 4. Fetch ALL active medicine batches (with location assignments)
	batchCursor, err := dbConfigs.MedicineBatchCollection.Find(ctx,
		bson.M{"locationId": bson.M{"$ne": ""}},
		options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("fetch batches: %w", err)
	}
	defer batchCursor.Close(ctx)
	var allBatches []dto.MedicineBatch
	if err = batchCursor.All(ctx, &allBatches); err != nil {
		return nil, err
	}

	// 5. Fetch ALL branch stock records (for quantity data)
	stockFilter := bson.M{}
	if branchId != "" {
		stockFilter["branchId"] = branchId
	}
	stockCursor, err := dbConfigs.BranchStockCollection.Find(ctx, stockFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch branch stock: %w", err)
	}
	defer stockCursor.Close(ctx)
	var allStocks []dto.BranchStock
	if err = stockCursor.All(ctx, &allStocks); err != nil {
		return nil, err
	}

	// 6. Collect unique medicine IDs to resolve names
	medIdSet := make(map[string]bool)
	for _, b := range allBatches {
		medIdSet[b.MedicineId] = true
	}
	medIds := make([]string, 0, len(medIdSet))
	for id := range medIdSet {
		medIds = append(medIds, id)
	}
	nameMap := make(map[string]string)
	if len(medIds) > 0 {
		nameMap, _ = DB_GetMedicineNamesByIDs(medIds)
	}

	// 7. Build lookup maps for in-memory joins
	// batchId → []BranchStockSlim
	stockByBatch := make(map[string][]dto.BranchStockSlim)
	for _, s := range allStocks {
		stockByBatch[s.BatchId] = append(stockByBatch[s.BatchId], dto.BranchStockSlim{
			BranchId:  s.BranchId,
			Quantity:  s.Quantity,
			Available: s.Quantity - s.ReservedQuantity,
		})
	}

	// locationId → []BatchSummary
	batchesByLocation := make(map[string][]dto.BatchSummary)
	for _, b := range allBatches {
		if b.LocationId == "" {
			continue
		}
		summary := dto.BatchSummary{
			BatchId:      b.BatchId,
			MedicineId:   b.MedicineId,
			MedicineName: nameMap[b.MedicineId],
			BatchNumber:  b.BatchNumber,
			ExpiryDate:   b.ExpiryDate,
			Status:       b.Status,
			BranchStocks: stockByBatch[b.BatchId],
		}
		if summary.BranchStocks == nil {
			summary.BranchStocks = []dto.BranchStockSlim{}
		}
		batchesByLocation[b.LocationId] = append(batchesByLocation[b.LocationId], summary)
	}

	// shelfId → []LocationWithBatches
	locationsByShelf := make(map[string][]dto.LocationWithBatches)
	for _, loc := range allLocations {
		batches := batchesByLocation[loc.LocationId]
		if batches == nil {
			batches = []dto.BatchSummary{}
		}
		lwb := dto.LocationWithBatches{
			LocationId:  loc.LocationId,
			Code:        loc.Code,
			Position:    loc.Position,
			Description: loc.Description,
			IsOccupied:  loc.IsOccupied,
			IsActive:    loc.IsActive,
			Batches:     batches,
		}
		locationsByShelf[loc.ShelfId] = append(locationsByShelf[loc.ShelfId], lwb)
	}

	// rackId → []ShelfWithLocations
	shelvesByRack := make(map[string][]dto.ShelfWithLocations)
	for _, shelf := range allShelves {
		locs := locationsByShelf[shelf.ShelfId]
		if locs == nil {
			locs = []dto.LocationWithBatches{}
		}
		swl := dto.ShelfWithLocations{
			ShelfId:     shelf.ShelfId,
			RackId:      shelf.RackId,
			Name:        shelf.Name,
			Description: shelf.Description,
			IsActive:    shelf.IsActive,
			Locations:   locs,
		}
		shelvesByRack[shelf.RackId] = append(shelvesByRack[shelf.RackId], swl)
	}

	// 8. Assemble the final tree
	result := make([]dto.RackWithShelves, 0, len(allRacks))
	for _, rack := range allRacks {
		shelves := shelvesByRack[rack.RackId]
		if shelves == nil {
			shelves = []dto.ShelfWithLocations{}
		}
		result = append(result, dto.RackWithShelves{
			RackId:      rack.RackId,
			Name:        rack.Name,
			Description: rack.Description,
			IsActive:    rack.IsActive,
			Shelves:     shelves,
		})
	}

	return result, nil
}

