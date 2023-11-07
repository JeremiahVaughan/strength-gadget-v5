package model

type ApiConfig struct {
	Name               string   `json:"name"`
	AllowedHttpMethods []string `json:"allowed_http_methods"`
	ImageUri           string   `json:"image_uri"`
	ConfigFilePath     string   `json:"config_file_path"`
}
