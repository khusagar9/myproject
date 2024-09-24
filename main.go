package main

/*import (
	"encoding/json"
	"fmt"
	"log"
	"myproject/api"
	"myproject/internal/noflyzone"
	"myproject/models"
)*/
/*
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
	var noFlyZones []noflyzone.NoFlyZone
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
	source := noflyzone.Point{Lat: 1.320000, Lon: 103.870000}
	destination := noflyzone.Point{Lat: 1.410000, Lon: 103.940000}

	//source := noflyzone.Point{Lat: 1.2500, Lon: 103.7000}
	//destination := noflyzone.Point{Lat: 1.4500, Lon: 103.7500}

	// Check if the path intersects any active no-fly zones
	if noflyzone.IsPathInNoFlyZone(source, destination, noFlyZones) {
		fmt.Println("The path intersects an active no-fly zone!")
	} else {
		fmt.Println("The path does not intersect any active no-fly zone.")
	}
}*/

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"myproject/config"
	"myproject/controller"
	"myproject/service"
	//"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/controller"
	//"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/log"
)

var (
	//GitCommit is the commit identifier injected at build time
	GitCommit string
)

func main() {
	// Getting configuration from config.*.json
	config.Load()
	log.Info("Loaded Configuration")

	defer func() { // recover the panic and exit -1
		if err := recover(); err != nil { //
			log.Error("panic: %v\n", err)
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(-1)
		}
	}()

	/*if GitCommit != "" {
		log.Info("GitCommit : [%s]", GitCommit)
	}*/

	backendID := "h3d-drone-emulator"
	log.Info("Starting [%s]", backendID)
	log.Info(".................................. H3D DroneSubSystem Successfully Started ..................................")

	//log.Info("Setting up Echo")
	//e := commonController.CreateEcho()
	e := echo.New()
	// log.Info("Initializing Services")
	service.InitService()

	controllers := make([]controller.IController, 0)
	controllers = append(controllers, controller.NewEmulatorSubsystem())

	for _, co := range controllers {
		if co != nil {
			co.Initialize(e)
		}
	}

	log.Info("Starting Echo")
	s := setupServer(e)

	log.Info("Setting up graceful shutdown management")
	gracefulShutdown(s, 5*time.Second, controllers)
}

// setupServer Setup server
func setupServer(e *echo.Echo) *http.Server {
	s := &http.Server{Addr: ":" + strconv.Itoa(*config.Get().ServerPort), Handler: e}
	done := make(chan bool)
	go func() {
		log.Info("Server exit: [%s]", s.ListenAndServe())
		done <- true
	}()
	return s
}

// Graceful shutdown of server when receive SIGTERM signal
func gracefulShutdown(hs *http.Server, timeout time.Duration, controllers []controller.IController) {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Dispose controllers
	log.Info("Disposing controllers")
	for _, co := range controllers {
		if co != nil {
			co.Dispose()
		}
	}

	log.Info("\nShutdown with timeout: %s\n", timeout)
	if err := hs.Shutdown(ctx); err != nil {
		log.Error("Error: %v\n", err)
	} else {
		log.Info("Server stopped")
	}
}
