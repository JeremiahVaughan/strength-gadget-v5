package config

import (
	"fmt"
	"os"
	"strengthgadget.com/m/v2/constants"
)

var AppUrl = "http://strengthgadget:8080" + constants.ApiPrefix

var NotificationEndpoint string

func InitConfig() error {
	NotificationEndpoint = os.Getenv("NOTIFICATION_ENDPOINT")
	if NotificationEndpoint == "" {
		return fmt.Errorf("error, must provide notification endpoint")
	}
	return nil
}
