package models

import "strconv"

// Struct for Drone Properties
type DroneH3D struct {
	DroneId          string     `json:"droneId"`
	DroneName        string     `json:"droneName"`
	CreatedBy        string     `json:"createdBy"`
	DbxId            string     `json:"dbxId"`
	Company          string     `json:"company"`
	SerialNo         string     `json:"serialNo"`
	CurrLat          float64    `json:"currLat"`
	CurrLong         float64    `json:"currLong"`
	CurrAltitude     float64    `json:"currAltitude"`
	CurrHeading      float64    `json:"currHeading"`
	DistanceFromHome float64    `json:"distanceFromHome"`
	GpsStatus        int        `json:"gpsStatus"`
	HomeLat          float64    `json:"homeLat"`
	HomeLong         float64    `json:"homeLong"`
	NetworkType      string     `json:"networkType"`
	BattLevel        float64    `json:"battLevel"`
	SignalStrength   string     `json:"signalStrength"`
	ErrorCode        int        `json:"errorCode"`
	Mission          Mission    `json:"mission"`
	DroneVideo       DroneVideo `json:"droneVideo"`
	Temperature      string     `json:"temperature"`
	TextualStatus    JSONData   `json:"textualStatus"`
	TimestampMs      int64      `json:"timestampMs"`
}

type DroneH3DResponse struct {
	DroneId     DroneH3DId       `json:"_id"`
	DroneName   string           `json:"name"`
	CameraId    string           `json:"serial_no"`
	CreateBy    string           `json:"create_by"`
	Dbx         Dbx              `json:"dbx"`
	Status      string           `json:"status"`
	Company     string           `json:"company"`
	TimestampMs DroneH3dCreateAt `json:"create_at"`
}

type DroneH3DId struct {
	DroneOid string `json:"$oid"`
}

type DroneH3dCreateAt struct {
	CreateAtDate int64 `json:"$date"`
}

type Dbx struct {
	Oid string `json:"$oid"`
}

func TransformDroneH3dFromDrone(drones []DroneH3D) []DroneH3DResponse {
	var dronesH3d []DroneH3DResponse
	for _, drone := range drones {
		droneH3d := DroneH3DResponse{
			DroneId: DroneH3DId{
				DroneOid: drone.DroneId,
			},
			DroneName: drone.DroneName,
			CreateBy:  drone.CreatedBy,
			Dbx: Dbx{
				Oid: drone.DbxId,
			},
			Status:   strconv.Itoa(drone.GpsStatus),
			CameraId: drone.SerialNo,
			Company:  drone.Company,
			TimestampMs: DroneH3dCreateAt{
				CreateAtDate: drone.TimestampMs,
			},
		}
		dronesH3d = append(dronesH3d, droneH3d)
	}
	return dronesH3d
}
