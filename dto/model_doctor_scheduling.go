package dto

import "time"

type DoctorWeeklySchedule struct {
	ID                string  `json:"id" bson:"_id,omitempty"`
	DoctorWeeklyScheduleID string `json:"doctorWeeklyScheduleId" bson:"doctorWeeklyScheduleId" binding:"required"`
	DoctorID          string  `json:"doctorId" bson:"doctorId" binding:"required"`
	DaysOfWeek        []int   `json:"daysOfWeek" bson:"daysOfWeek" binding:"required"`
	DefaultStartTime  *string `json:"defaultStartTime,omitempty" bson:"defaultStartTime,omitempty"`
	IsActive          bool    `json:"isActive" bson:"isActive"`
}

type DoctorAvailability struct {
	ID                 string    `json:"id" bson:"_id,omitempty"`
	DoctorAvailabilityID string `json:"doctorAvailabilityId" bson:"doctorAvailabilityId" binding:"required"`
	DoctorID           string    `json:"doctorId" bson:"doctorId" binding:"required"`
	Date               string    `json:"date" bson:"date" binding:"required"` // "2026-03-26"
	IsAvailable        bool      `json:"isAvailable" bson:"isAvailable"`
	EstimatedStartTime *string   `json:"estimatedStartTime,omitempty" bson:"estimatedStartTime,omitempty"`
	MaxPatients        *int      `json:"maxPatients,omitempty" bson:"maxPatients,omitempty"`
	Notes              *string   `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" bson:"updatedAt"`
}

type AvailableDate struct {
	Date             string  `json:"date"`
	DayOfWeek        int     `json:"dayOfWeek"`
	DayName          string  `json:"dayName"`
	DefaultStartTime *string `json:"defaultStartTime"`
}

type AvailableDateResponse struct {
	AvailableDates []AvailableDate `json:"availableDates"`
}
