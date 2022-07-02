package features

import (
	"api_chat/config"
	"api_chat/models"
	"encoding/json"
	"strings"
)

// ValidateEvent ensures data is a valid JSON representation of Chat Event and can be parsed as such
func ValidateEvent(data []byte) (models.ChatEvent, error) {
	var evt models.ChatEvent

	if err := json.Unmarshal(data, &evt); err != nil {
		return evt, &config.APIError{Code: 303}
	}

	if evt.User == "" {
		return evt, &config.APIError{Code: 303, Field: "name"}
	} else if evt.Msg == "" && strings.ToLower(evt.EventType) == models.Broadcast {
		return evt, &config.APIError{Code: 303, Field: "msg"}
	}

	return evt, nil
}
