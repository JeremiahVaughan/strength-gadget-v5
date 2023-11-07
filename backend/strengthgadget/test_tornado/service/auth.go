package service

import (
	"fmt"
	"math/rand"
	"strengthgadget.com/m/v2/test_tornado/constants"
	"strings"
)

func GetValidEmail() string {
	validEmailParts := strings.Split(constants.ValidEmail, "@")
	return fmt.Sprintf("%s%d@%s", validEmailParts[0], rand.Intn(4294967294), validEmailParts[1])
}
