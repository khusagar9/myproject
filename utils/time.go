package utils

import (
	"myproject/models"
)

// IsZoneActive checks if the zone is active based on the current time
func IsZoneActive(zone models.NoFlyZone, currentTime int64) bool {
	return currentTime >= zone.StartTime && currentTime <= zone.EndTime
}
