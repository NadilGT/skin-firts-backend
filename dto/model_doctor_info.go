package dto

import "time"

type DoctorInfoModel struct {
	DoctorID   string   `json:"doctor_id" bson:"doctor_id"`
	Name       string   `json:"name" bson:"name"`
	Experience int      `json:"experience" bson:"experience"`
	Focus      string   `json:"focus" bson:"focus"`
	FocusId    string   `json:"focus_id" bson:"focus_id"`
	Special    string   `json:"special" bson:"special"`
	Starts     int      `json:"starts" bson:"starts"`
	Messages   int      `json:"messages" bson:"messages"`
	Date       string   `json:"date" bson:"date"`
	Profile    string   `json:"profile" bson:"profile"`
	Career     string   `json:"career" bson:"career"`
	Highlights string   `json:"highlights" bson:"highlights"`
	Favorite   bool     `json:"favorite" bson:"favorite"`
	ProfilePic string   `json:"profile_pic" bson:"profile_pic"`
	// Multi-branch support: a doctor can work at multiple branches
	BranchIds  []string  `json:"branchIds,omitempty" bson:"branchIds,omitempty"`
	Status     string    `json:"status,omitempty" bson:"status,omitempty"` // ACTIVE / INACTIVE
	UpdatedAt  time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

// SearchDoctorInfoQuery represents the query parameters for searching doctor info.
type SearchDoctorInfoQuery struct {
	Query   string `json:"query" query:"query"`
	Focus   string `json:"focus" query:"focus"`
	FocusId string `json:"focusId" query:"focusId"`
	Special string `json:"special" query:"special"`
	Page    int    `json:"page" query:"page"`
	Limit   int    `json:"limit" query:"limit"`
}

// DoctorInfoSearchResponse represents the paginated response for a doctor info search.
type DoctorInfoSearchResponse struct {
	Data       []DoctorInfoModel `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"totalPages"`
}
