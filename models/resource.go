package models

// Struct for Resource

type Resource struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	FirstName     *string `json:"firstName"`
	LastName      *string `json:"lastName"`
	UserId        *string `json:"userId"`
	IsVehicle     bool    `json:"isVehicle"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	BaseLatitude  float64 `json:"baseLatitude"`
	BaseLongitude float64 `json:"baseLongitude"`
}

type ResourceLocation struct {
	ResourceId     string  `json:"resourceId"`
	Location       string  `json:"location"`
	Altitude       float64 `json:"altitude"`
	IsExternal     bool    `json:"isExternal"`
	IsVehicle      bool    `json:"isVehicle"`
	GenTimestampMs int64   `json:"genTimestampMs"`
	TimestampMs    int64   `json:"timestampMs"`
}
