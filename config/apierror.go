package server

import (
	"fmt"
	"strconv"
)

type APIError struct {
	Code  int    `json:"code,omitempty"`
	Msg   string `json:"error,omitempty"`
	Field string `json:"field,omitempty"`
}

type Outcome struct {
	Status bool      `json:"status"`
	Error  *APIError `json:"error,omitempty"`
}

func (e *APIError) SetMsg() {
	switch e.Code {
	case 101:
		e.Msg = "Room error: Room not found"
	case 102:
		e.Msg = "Room error: Duplicate room"
	case 103:
		e.Msg = "Room error: Invalid JSON"
	case 104:
		e.Msg = "Room error: Unauthorized operation"
	case 105:
		e.Msg = "Room error: Invalid content"
	case 201:
		e.Msg = "Client error: User not found"
	case 202:
		e.Msg = "Client error: Duplicate username"
	case 203:
		e.Msg = "Client error: Invalid JSON"
	case 204:
		e.Msg = "Client error: Unauthorized operation"
	case 301:
		e.Msg = "Could not establish session"
	case 303:
		e.Msg = "Invalid JSON"
	case 304:
		e.Msg = "Unauthorized operation"
	case 305:
		e.Msg = "Unsupported client device"
	case 401:
		e.Msg = "Token error: Invalid signature"
	case 402:
		e.Msg = "Token error: Unauthorized signing method"
	case 403:
		e.Msg = "Token error: Invalid token"
	default:
		e.Msg = "Unknown error: " + e.Msg
	}
}

func (e *APIError) Error() string {
	e.SetMsg()
	if e.Field != "" {
		return fmt.Sprintf("{\"error\": \"%s\", \"code\": %d, \"field\": \"%s\"}", e.Msg, e.Code, e.Field)
	}
	return fmt.Sprintf("{\"error\": \"%s\", \"code\": %d}", e.Msg, e.Code)
}

func isInt(titleorID string) int {
	if id, err := strconv.Atoi(titleorID); err == nil {
		return id
	}
	return -1
}
