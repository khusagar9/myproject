package models

import (
	"log"
	"strconv"
	"strings"
	"time"
)

// Struct for Drone Properties
type DroneStatus struct {
	ResourceId       string   `json:"resourceId"`
	Speed            float32  `json:"speed"`
	BatteryLevel     float32  `json:"batteryLevel"`
	Heading          float32  `json:"heading"`
	DistanceFromHome float32  `json:"distanceFromHome"`
	GpsStatus        string   `json:"gpsStatus"`
	SignalStrength   string   `json:"signalStrength"`
	Temperature      float32  `json:"temperature"`
	TimestampMs      int64    `json:"timestampMs"`
	GenTimestampMs   int64    `json:"genTimestampMs"`
	TextualStatus    JSONData `json:"textualStatus"`
}

func TransformDroneStatusFromH3dStatus(h3dDrone DroneH3dStatus, resourceId string) DroneStatus {
	speed64, err := strconv.ParseFloat(strings.Replace(h3dDrone.DroneSpeed, " mph", "", 1), 32)
	if err != nil {
		log.Println("TransformDroneStatusFromH3dStatus Error:", err)
	}
	var speed = float32(speed64)

	batt64, err := strconv.ParseFloat(h3dDrone.BattLevel, 32)
	if err != nil {
		log.Println("TransformDroneStatusFromH3dStatus Error:", err)
	}
	var battLevel = float32(batt64)

	dist64, err := strconv.ParseFloat(strings.Replace(h3dDrone.DistanceFromHome, " M", "", 1), 32)
	if err != nil {
		log.Println("TransformDroneStatusFromH3dStatus Error:", err)
	}
	var dist = float32(dist64)

	temp64, err := strconv.ParseFloat(h3dDrone.Temperature, 32)
	if err != nil {
		log.Println("TransformDroneStatusFromH3dStatus Error:", err)
	}
	var temp = float32(temp64)

	droneStatus := DroneStatus{
		ResourceId:       resourceId,
		Speed:            speed,
		BatteryLevel:     battLevel,
		Heading:          float32(h3dDrone.CurrHeading),
		DistanceFromHome: dist,
		GpsStatus:        strconv.Itoa(h3dDrone.GpsStatus),
		SignalStrength:   h3dDrone.SignalStrength,
		Temperature:      temp,
		TimestampMs:      time.Now().UnixNano() / int64(time.Millisecond),
		GenTimestampMs:   time.Now().UnixNano() / int64(time.Millisecond),
		// Sample TextuaStatus
		TextualStatus: h3dDrone.TextualStatus,
	}
	return droneStatus
}
