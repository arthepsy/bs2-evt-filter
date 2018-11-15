package main

import (
	"bs2-evt-filter/internal/pkg/config"
	"bs2-evt-filter/pkg/biostar2"
)

func filterEvent(filter *config.FilterConf, e *biostar2.Event) bool {
	if len(filter.EventTypeCodes) > 0 {
		_, ok := filter.EventTypeCodes[e.EventType.Code]
		if !ok {
			return false
		}
	}
	if len(filter.DeviceIDs) > 0 {
		_, ok := filter.DeviceIDs[e.Device.ID]
		if !ok {
			return false
		}
	}
	return true
}
