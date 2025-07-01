package models

type User struct {
	ID           uint   `json:"id" example:"1"`
	Username     string `json:"username" example:"john_doe"`
	Password     string `json:"password"`
	RefreshToken string `json:"-"`
}
