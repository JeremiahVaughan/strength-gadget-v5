package model

type HealthResponse struct {
	Status     string `json:"status"`
	AppVersion string `json:"appVersion"`
}
