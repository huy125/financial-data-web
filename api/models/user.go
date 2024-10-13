package model

// User represents user information
type User struct {
	ID		int `json:"id"`
	Username 	string `json:"username"`
	Hash 		string `json:"hash"`
}
