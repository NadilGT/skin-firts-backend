package dto

type DoctorWeeklySchedule struct {
	ID                string  `json:"id" bson:"_id,omitempty"`
	DoctorWeeklyScheduleID string `json:"doctorWeeklyScheduleId" bson:"doctorWeeklyScheduleId" binding:"required"`
	DoctorID          string  `json:"doctorId" bson:"doctorId" binding:"required"`
	DaysOfWeek        []int   `json:"daysOfWeek" bson:"daysOfWeek" binding:"required"`
	DefaultStartTime  *string `json:"defaultStartTime,omitempty" bson:"defaultStartTime,omitempty"`
	IsActive          bool    `json:"isActive" bson:"isActive"`
	BranchId          string  `json:"branchId" bson:"branchId"`
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
type CheckAvailabilityResponse struct {
	IsAvailable bool   `json:"isAvailable"`
	Message     string `json:"message"`
}
