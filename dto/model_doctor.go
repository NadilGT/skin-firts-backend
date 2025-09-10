package dto

type Doctor struct {
    DoctorName   string   `json:"doctorName" bson:"doctorName"`
    Specialty    string   `json:"specialty" bson:"specialty"`
    ProfilePic   string   `json:"profilePic" bson:"profilePic"`
    Rating       float64  `json:"rating" bson:"rating"`
    ReviewCount  int      `json:"reviewCount" bson:"reviewCount"`
}