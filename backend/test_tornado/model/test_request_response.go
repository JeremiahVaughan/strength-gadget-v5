package model

import "net/http"

type TestRequestResponse struct {
	ResponseCode    int
	Cookie          *http.Cookie
	ResponsePayload any
}
