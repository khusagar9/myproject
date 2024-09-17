package models

// Coordinate defines latitude and longitude
type Coordinate struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

// NoFlyZone represents a restricted zone
type NoFlyZone struct {
	ID          string           `json:"id"`
	Coordinates [][][]Coordinate `json:"geometry_coordinates"`
	StartTime   int64            `json:"activation_start_timestamp_ms"`
	EndTime     int64            `json:"activation_end_timestamp_ms"`
	AreaPath    string           `json:"area_path"`
}
