package dto

import "time"

type DoctorScheduleModel struct {
	DoctorName string    `json:"doctorName" bson:"doctorName"`
	Date       time.Time `json:"date" bson:"date"`
	TimeSlots  []string  `json:"timeSlots" bson:"timeSlots"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
}

type CreateDoctorScheduleRequest struct {
	DoctorName string    `json:"doctorName" binding:"required"`
	Date       time.Time `json:"date" binding:"required"`
	TimeSlots  []string  `json:"timeSlots" binding:"required"`
}


type DoctorScheduleResponse struct {
	DoctorName string                     `json:"doctorName"`
	Schedules  map[string][]string        `json:"schedules"`
}