package models

type Notification struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
