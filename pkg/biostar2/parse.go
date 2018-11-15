package biostar2

import (
	"encoding/json"
)

func ParseResponse(data []byte) (*Response, bool) {
	var w ResponseWrapper
	err := json.Unmarshal(data, &w)
	if err == nil && len(w.Response.Code) > 0 {
		return &w.Response, true
	} else {
		return nil, false
	}
}

func ParseEvent(data []byte) (*Event, bool) {
	var w EventWrapper
	err := json.Unmarshal(data, &w)
	if err == nil && len(w.Event.Device.ID) > 0 {
		return &w.Event, true
	} else {
		return nil, false
	}
}
