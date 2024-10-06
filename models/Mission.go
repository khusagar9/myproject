package models

// Struct for Drone Mission Properties
type Mission struct {
	//ApiKey     *string     `json:"API_KEY"`
	Waypoints  [][]float64 `json:"waypoints"`
	NewMission NewMission  `json:"newMission"`
	DbxId      *string     `json:"dbx_id"`
	DroneId    *string     `json:"drone_id"`
}

type MissionCommand struct {
	OperationId    *string     `json:"operationId"`
	ResourceId     *string     `json:"resourceId"`
	MissionId      *string     `json:"missionId"`
	MissionName    *string     `json:"missionName"`
	MissionType    *string     `json:"missionType"`
	Waypoints      [][]float64 `json:"waypoints"`
	GenTimestampMs int64       `json:"genTimestampMs"`
}
type NewMission struct {
	MissionName string `json:"missionName"`
}
