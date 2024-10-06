package models

// Struct for Drone Properties
type DroneH3dStatus struct {
	Altitude         int      `json:"altitude"`
	BattLevel        string   `json:"battery_level"`
	DistanceFromHome string   `json:"distance_home"`
	DroneSpeed       string   `json:"drone_speed"`
	DronesPosition   string   `json:"drones_position"`
	GpsStatus        int      `json:"gps_status"`
	CurrHeading      int      `json:"heading"`
	HomePosition     string   `json:"home_position"`
	NetworkType      string   `json:"network_type"`
	SignalStrength   string   `json:"signal_strength"`
	Temperature      string   `json:"temperature"`
	TextualStatus    JSONData `json:"textualStatus"`
}

type DroneVideo struct {
	Link1 string `json:"link1"`
	Link2 string `json:"link2"`
}
