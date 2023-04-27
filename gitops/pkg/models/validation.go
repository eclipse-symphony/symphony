package models

type Error struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type ValidateResponse struct {
	Status string  `json:"status"`
	Error  *Error  `json:"error,omitempty"`
}