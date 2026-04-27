package dto

type DoctorWeeklySchedule struct {
	ID                     string  `json:"id" bson:"_id,omitempty"`
	DoctorWeeklyScheduleID string  `json:"doctorWeeklyScheduleId" bson:"doctorWeeklyScheduleId" binding:"required"`
	DoctorID               string  `json:"doctorId" bson:"doctorId" binding:"required"`
	DaysOfWeek             []int   `json:"daysOfWeek" bson:"daysOfWeek" binding:"required"`
	DefaultStartTime       *string `json:"defaultStartTime,omitempty" bson:"defaultStartTime,omitempty"`
	MaxPatients            int     `json:"maxPatients" bson:"maxPatients"` // 0 = unlimited
	IsActive               bool    `json:"isActive" bson:"isActive"`
	BranchId               string  `json:"branchId" bson:"branchId"`
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

// DoctorDailyCapacity tracks how many patients are booked vs the max
// for a specific doctor + branch + date. Lazily initialized on first booking.
type DoctorDailyCapacity struct {
	DoctorDailyCapacityId string `json:"doctorDailyCapacityId" bson:"doctorDailyCapacityId"`
	DoctorID              string `json:"doctorId" bson:"doctorId"`
	BranchId              string `json:"branchId" bson:"branchId"`
	Date                  string `json:"date" bson:"date"` // UTC date "2026-04-29"
	Booked                int    `json:"booked" bson:"booked"`
	Max                   int    `json:"max" bson:"max"`
}
