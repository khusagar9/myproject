package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"myproject/api"
	"myproject/config"
	"myproject/internal/noflyzone"
	"myproject/models"
	"strconv"
	"strings"
	"time"
)

var applicationConfig config.AppConfig
var noFlyZones []noflyzone.NoFlyZone
var source noflyzone.Point
var destination noflyzone.Point

//var httpClient *http.Client

func InitService() error {
	applicationConfig = config.Get()
	//getPath()
	return nil
	//httpClient = commonHttp.CreateHttpClient(nil)
}
func getNoFlyZone() {
	client := api.NewClient()

	requestBody := models.RequestBody{
		CeInstance: models.CeInstance{
			TemplateID: models.TemplateID{
				Equals: []string{"no_fly_zone"},
			}},
		Offset:    0,
		Limit:     1000,
		Order:     "creation_date:asc",
		WithTotal: true,
	}

	response, err := client.Post(*config.Get().NoFlyZoneEndPoint, requestBody)
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}

	var result models.ApiResponse
	if err := json.Unmarshal(response, &result); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}
	//var noFlyZones []noflyzone.NoFlyZone
	for _, instance := range result.CeInstances {
		var polygon []noflyzone.Point

		if outer, ok := instance.Geometry.Coordinates.([]interface{}); ok {
			if middle, ok := outer[0].([]interface{}); ok {
				if inner, ok := middle[0].([]interface{}); ok {
					for _, coordinate := range inner {
						if firstCoordinate, ok := coordinate.([]interface{}); ok {
							// Extract latitude and longitude from the first coordinate
							if len(firstCoordinate) >= 2 {
								longitude := firstCoordinate[0].(float64)
								latitude := firstCoordinate[1].(float64)
								polygon = append(polygon, noflyzone.Point{Lat: latitude, Lon: longitude})
							}
						}
					}
				}
			}
		}

		noFlyZone := noflyzone.NoFlyZone{
			Polygon:   polygon,
			StartTime: instance.Data.ActivationStart.TimestampMs,
			EndTime:   instance.Data.ActivationEnd.TimestampMs,
		}
		noFlyZones = append(noFlyZones, noFlyZone)
	}

	// Example source and destination points in Singapore

}

func GetSourceDestinationPoints(coordinates string) (noflyzone.Point, noflyzone.Point, error) {
	parts := strings.Split(coordinates, ":")

	// Split the first part to get the source lat/lon
	sourceCoords := strings.Split(parts[0], ",")
	sourceLat, _ := strconv.ParseFloat(sourceCoords[0], 64)
	sourceLon, _ := strconv.ParseFloat(sourceCoords[1], 64)
	source = noflyzone.Point{Lat: sourceLat, Lon: sourceLon}

	// Split the second part to get the destination lat/lon
	destinationCoords := strings.Split(parts[1], ",")
	destinationLat, _ := strconv.ParseFloat(destinationCoords[0], 64)
	destinationLon, _ := strconv.ParseFloat(destinationCoords[1], 64)
	destination = noflyzone.Point{Lat: destinationLat, Lon: destinationLon}

	fmt.Println(source.Lat)
	fmt.Println(destination.Lat)
	return source, destination, nil

}

func GetPath(rc *models.RequestContext, query string) (float64, error) {
	getNoFlyZone()
	source, destination, nil := GetSourceDestinationPoints(query)
	//source := noflyzone.Point{Lat: 1.320000, Lon: 103.870000}
	//destination := noflyzone.Point{Lat: 1.410000, Lon: 103.940000}

	// 1.320000,103.870000:1.410000,103.940000
	//source := noflyzone.Point{Lat: 1.2500, Lon: 103.7000}
	//destination := noflyzone.Point{Lat: 1.4500, Lon: 103.7500}

	// Check if the path intersects any active no-fly zones
	if noflyzone.IsPathInNoFlyZone(source, destination, noFlyZones) {
		fmt.Println("The path intersects an active no-fly zone!")
		return 0, errors.New("Path has intersect flyzone")
	} else {
		fmt.Println("The path does not intersect any active no-fly zone.")
		distance := haversineDistance(source, destination)
		distanceInMeters := distance * 1000
		return distanceInMeters, nil
	}

}

func haversineDistance(p1, p2 noflyzone.Point) float64 {
	const R = 6371.0 // Earth radius in kilometers

	// Convert latitude and longitude from degrees to radians
	lat1 := degreesToRadians(p1.Lat)
	lon1 := degreesToRadians(p1.Lon)
	lat2 := degreesToRadians(p2.Lat)
	lon2 := degreesToRadians(p2.Lon)

	// Haversine formula
	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	return R * c
}

// degreesToRadians converts degrees to radians.
func degreesToRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func GetWaypoints(source, destination noflyzone.Point, startTime, endTime time.Time, numWaypoints int) []struct {
	Point noflyzone.Point
	Time  time.Time
} {
	// Calculate total distance and total time
	//totalDistance := haversineDistance(source, destination)
	totalTime := endTime.Sub(startTime)

	// Calculate time per waypoint
	timeStep := totalTime / time.Duration(numWaypoints-1)

	// Generate waypoints with timestamps
	waypoints := make([]struct {
		Point noflyzone.Point
		Time  time.Time
	}, numWaypoints)

	for i := 0; i < numWaypoints; i++ {
		fraction := float64(i) / float64(numWaypoints-1) // Progress fraction [0, 1]
		waypoints[i].Point = interpolate(source, destination, fraction)
		waypoints[i].Time = startTime.Add(timeStep * time.Duration(i))
	}

	return waypoints
}

// Linear interpolation between two points.
func interpolate(p1, p2 noflyzone.Point, fraction float64) noflyzone.Point {
	return noflyzone.Point{
		Lat: p1.Lat + (p2.Lat-p1.Lat)*fraction,
		Lon: p1.Lon + (p2.Lon-p1.Lon)*fraction,
	}
}

func interpolate2(p1, p2 noflyzone.Point, fraction float64) models.Point {
	return models.Point{
		Latitude:  p1.Lat + (p2.Lat-p1.Lat)*fraction,
		Longitude: p1.Lon + (p2.Lon-p1.Lon)*fraction,
		Altitude:  0, // Assuming constant altitude for now
	}
}

func GetWaypoints2(source, destination noflyzone.Point, startTime, endTime time.Time, numWaypoints int) []models.Point {
	// Calculate total distance and total time
	//totalDistance := haversine(source, destination)
	totalTime := endTime.Sub(startTime)

	// Calculate time per waypoint
	timeStep := totalTime / time.Duration(numWaypoints-1)

	// Generate waypoints with timestamps
	waypoints := make([]models.Point, numWaypoints)

	for i := 0; i < numWaypoints; i++ {
		fraction := float64(i) / float64(numWaypoints-1) // Progress fraction [0, 1]
		waypoint := interpolate2(source, destination, fraction)
		waypoint.Time = startTime.Add(timeStep * time.Duration(i))
		waypoints[i] = waypoint
	}

	return waypoints
}
