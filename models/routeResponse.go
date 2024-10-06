package models

import "time"

// RouteResponse represents the entire response containing routes.
type RouteResponse struct {
	Routes []Route `json:"routes"`
}

// Route contains summary information and legs of a route.
type Route struct {
	Summary RouteSummary `json:"summary"`
	Legs    []Leg        `json:"legs"`
}

// RouteSummary contains information summarizing the route.
type RouteSummary struct {
	LengthInMeters                   int             `json:"lengthInMeters"`
	TravelTimeInSeconds              int             `json:"travelTimeInSeconds"`
	TrafficDelayInSeconds            int             `json:"trafficDelayInSeconds"`
	TrafficLengthInMeters            int             `json:"trafficLengthInMeters"`
	DepartureTime                    time.Time       `json:"departureTime"`
	ArrivalTime                      time.Time       `json:"arrivalTime"`
	ClearanceRequired                bool            `json:"clearanceRequired"`
	RemainingOperationTimeAtLocation int             `json:"remainingOperationTimeAtLocationInSeconds"`
	ClearanceZones                   []ClearanceZone `json:"clearanceZones"`
}

// Leg represents each leg of the journey with its own summary and points.
type Leg struct {
	Summary LegSummary `json:"summary"`
	Points  []Point    `json:"points"`
}

// LegSummary contains information summarizing a leg of a route.
type LegSummary struct {
	LengthInMeters                   int             `json:"lengthInMeters"`
	TravelTimeInSeconds              int             `json:"travelTimeInSeconds"`
	TrafficDelayInSeconds            int             `json:"trafficDelayInSeconds"`
	TrafficLengthInMeters            int             `json:"trafficLengthInMeters"`
	DepartureTime                    time.Time       `json:"departureTime"`
	ArrivalTime                      time.Time       `json:"arrivalTime"`
	ClearanceRequired                bool            `json:"clearanceRequired"`
	RemainingOperationTimeAtLocation int             `json:"remainingOperationTimeAtLocationInSeconds"`
	ClearanceZones                   []ClearanceZone `json:"clearanceZones"`
}

// ClearanceZone represents a no-fly zone with entry and exit times.
type ClearanceZone struct {
	ID        string    `json:"id"`
	EntryTime time.Time `json:"entryTime"`
	ExitTime  time.Time `json:"exitTime"`
}

// Point represents a specific point on the map with latitude, longitude, altitude, and time.
type Point struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Time      time.Time `json:"time"`
	//Altitude  float64   `json:"altitude"`
}
