package dto

type DoctorInfoModel struct {
	Name       string   `json:"name" bson:"name"`
	Experience int      `json:"experience" bson:"experience"`
	Focus      string   `json:"focus" bson:"focus"`
	Special    string   `json:"special" bson:"special"`
	Starts     int      `json:"starts" bson:"starts"`
	Messages   int      `json:"messages" bson:"messages"`
	Date       string   `json:"date" bson:"date"`
	Profile    string   `json:"profile" bson:"profile"`
	Career     string   `json:"career" bson:"career"`
	Highlights string `json:"highlights" bson:"highlights"`
}
