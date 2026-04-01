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
