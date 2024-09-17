package main

import (
	"encoding/json"
	"fmt"
	"log"
	"myproject/api"
	"myproject/models"
	"myproject/utils"
	"time"
)

func main() {
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

	response, err := client.Post("https://pilot.sdpcore.apps.thalesdigital.io/custom_entity/v0/internal/instances/search", requestBody)
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}

	var result models.ApiResponse
	if err := json.Unmarshal(response, &result); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}
	// Loop through CeInstances and convert to NoFlyZone
	var noFlyZones []models.NoFlyZone
	for _, instance := range result.CeInstances {
		fmt.Printf("ZoneID: %s\n", instance.ID)
		fmt.Printf("Geometry Coordinates: %v\n", instance.Geometry.Coordinates)
		fmt.Printf("Area Path: %s\n", instance.AreaPath)
		fmt.Printf("ActivationStartTimestampMs: %v\n", instance.Data.ActivationStart.TimestampMs)
		fmt.Printf("ActivationEndTimestampMs: %v\n", instance.Data.ActivationEnd.TimestampMs)

		// Convert the coordinates (assuming they're in the correct format)
		var coordinates [][][]models.Coordinate
		if coordSlice, ok := instance.Geometry.Coordinates.([][][]float64); ok {
			for _, polygon := range coordSlice {
				var polygonCoords []models.Coordinate
				for _, coord := range polygon {
					polygonCoords = append(polygonCoords, models.Coordinate{
						Lon: coord[0],
						Lat: coord[1],
					})
				}
				coordinates = append(coordinates, [][]models.Coordinate{polygonCoords})
				//coordinates = append(coordinates, polygonCoords)
			}
		}

		// Create a new NoFlyZone
		noFlyZone := models.NoFlyZone{
			ID:          instance.ID,
			Coordinates: coordinates,
			StartTime:   instance.Data.ActivationStart.TimestampMs,
			EndTime:     instance.Data.ActivationEnd.TimestampMs,
			AreaPath:    instance.AreaPath,
		}

		// Append the new noFlyZone to the slice
		noFlyZones = append(noFlyZones, noFlyZone)
	}

	// At this point, noFlyZones contains the populated data from CeInstances
	fmt.Println("NoFlyZones populated from CeInstances:", noFlyZones)
	// Example source and destination coordinates
	source := models.Coordinate{Lat: 1.3000, Lon: 103.8000}
	dest := models.Coordinate{Lat: 1.4000, Lon: 103.9000}

	// Check for no-fly zone intersection
	currentTime := time.Now().UnixMilli() // Current timestamp in ms
	for _, zone := range noFlyZones {
		if utils.IsZoneActive(zone, currentTime) {
			if utils.IsPathIntersectingZone(source, dest, zone) {
				fmt.Printf("Path intersects no-fly zone ID: %s\n", zone.ID)
			}
		}
	}
}
