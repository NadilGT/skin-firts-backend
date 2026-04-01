package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"
)

// DB_SaveReport inserts a new report into the reports collection.
func DB_SaveReport(report dto.ReportModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := dbConfigs.ReportCollection.InsertOne(ctx, report)
	if err != nil {
		return fmt.Errorf("failed to save report to database: %w", err)
	}
	return nil
}

// DB_GetReportsByPatientID fetches all reports associated with a patient by their Firebase UID.
func DB_GetReportsByPatientID(patientID string) ([]dto.ReportModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reports []dto.ReportModel
	cursor, err := dbConfigs.ReportCollection.Find(ctx, map[string]string{"patientId": patientID})
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &reports); err != nil {
		return nil, fmt.Errorf("failed to decode reports: %w", err)
	}

	return reports, nil
}
