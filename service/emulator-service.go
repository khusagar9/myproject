package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"h3d-drone-emulator/api"
	"h3d-drone-emulator/config"
	"h3d-drone-emulator/models"
	restrictedZone "h3d-drone-emulator/util"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	commonHttp "gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/http"
	"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/log"
	commonS3 "gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/s3"
)

var applicationConfig config.AppConfig
var drones []models.DroneH3D

var source restrictedZone.Point
var destination restrictedZone.Point

// var missionStatuses []int = []int{0, 0, 0}

// var startingLat float64
// var startingLong float64

var textualStatuses []models.JSONData = []models.JSONData{{"Dbx_id": "123ABC", "Country": "Temasek", "network_type": "4G", "home_position": "1.334944,103.737051"}, {"Dbx_id": "1234567", "Country": "Singapore"}, {}}

// var signalStrengths []string = []string{"Excellent", "Good", "Fair", "Bad"}
var dbxIds []string = []string{"603f1c1dbf35eba727ee6c3a", "602c8187519cc032dd65e739", "604743f99ecead5820683080"}

var httpClient *http.Client

var resources []models.Resource
var resourceStatusMap map[string]string
var patrolStatus = "PATROL"
var missionStatus = "MISSION"
var returnToBaseStatus = "RETURN_TO_BASE"
var missionChanMap map[string]chan string
var patrolStopChan chan int

var simuMap map[string][][]float64
var simuMapMutex sync.RWMutex

var locChan = make(chan models.Resource, 10)

var s3Client *s3.S3

func randomInt(min int, max int) int {
	return rand.Intn(max-min) + min
}

func InitService() {
	applicationConfig = config.Get()
	httpClient = commonHttp.CreateHttpClient(nil)
	s3, err := commonS3.CreateClient(config.Get().S3Config)
	if err != nil {
		panic("Could not create S3 client: " + err.Error())
	}
	s3Client = s3
	// droneIds := strings.Split(*applicationConfig.DroneIds, ",")
	// simulateRealH3D = *applicationConfig.SimulateRealH3D
	// Publishes location of Drone every 5 seconds
	//go publishLocation(ctx)
	// Simulate Movement for H3D Drone, not in use
	// if !simulateRealH3D {
	// 	go simulateMovement()
	// }
	// parse resources.json to the resources
	if s3Client != nil {
		log.Info("Getting resources.json from S3")
		err := commonS3.ReadJsonFile(s3Client, "sdp-rms-external-simulator", "pilot_resources.json", &resources)
		if err != nil {
			log.Error("Could not get resources.json: %s", err.Error())
		}
	}
	log.Info("resources %#v", resources)
	resourceStatusMap = make(map[string]string)

	// droneIds := make([]string, 0)
	for i, res := range resources {
		resourceStatusMap[res.ID] = patrolStatus
		if res.Type == "DRONE" {
			drone := models.DroneH3D{
				DroneId:          res.ID,
				DroneName:        "H3d Drone " + strconv.Itoa(i+1),
				CreatedBy:        "H3d",
				DbxId:            dbxIds[i%3],
				Company:          "H3d",
				SerialNo:         "H3D00" + strconv.Itoa(i+1),
				CurrLat:          res.BaseLatitude,
				CurrLong:         res.BaseLongitude,
				CurrAltitude:     0,
				CurrHeading:      float64(randomInt(0, 360)),
				DistanceFromHome: float64(randomInt(0, 51)),
				GpsStatus:        randomInt(5, 7),
				HomeLat:          res.BaseLatitude,
				HomeLong:         res.BaseLongitude,
				BattLevel:        float64(randomInt(50, 101)),
				SignalStrength:   *applicationConfig.SignalStrength,
				Temperature:      strconv.Itoa(*applicationConfig.Temperature),
				TextualStatus:    textualStatuses[i%3],
				ErrorCode:        0,
				TimestampMs:      time.Now().UnixNano(),
				DroneVideo: models.DroneVideo{
					Link1: `"flv: "http://localhost:8000/live/N2_cctv2.flv"`,
					Link2: `m3u8: "http://localhost:8000/live/N2_cctv2.m3u8"`,
				},
				Mission: models.Mission{
					//MissionId:   0,
					NewMission: models.NewMission{
						MissionName: "",
					},
					Waypoints: [][]float64{},
				},
			}
			drones = append(drones, drone)
		}
	}

	// Printing out Drone Information using regex
	// log.Info("****************** H3D Drones Initiatied: ******************")
	// log.Info("%#v", drones)

	// parse simulation files under files folder
	simuMap = make(map[string][][]float64)
	chunks := chunkResources(resources)
	timeStart := time.Now()
	var wg sync.WaitGroup
	for _, chunk := range chunks {
		for _, res := range chunk {
			wg.Add(1)
			go func(res models.Resource) {
				if s3Client != nil {
					log.Info("Processing %s", res.ID)
					key := res.ID + ".csv"
					simulation, err := commonS3.ReadBytes(s3Client, "sdp-rms-external-simulator", key)
					if err != nil {
						log.Error("Could not get %s: %s", key, err.Error())
					}
					waypoints := getWayPoints(simulation)
					simuMapMutex.Lock()
					simuMap[res.ID] = waypoints
					simuMapMutex.Unlock()
					wg.Done()
				}
			}(res)
		}
		wg.Wait()
	}
	log.Info("Time taken: %v", time.Since(timeStart))
	/*
		for k, v := range simuMap {
			log.Info(k)
			for i, point := range v {
				log.Info("%d %f %f", i, point[0], point[1])
			}
		}
	*/
	go updateLocations()
	patrolStopChan = make(chan int)
	for k := range simuMap {
		isVehicle := false
		for _, res := range resources {
			if res.ID == k {
				isVehicle = res.IsVehicle
				break
			}
		}
		if len(simuMap[k]) > 0 {
			// log.Info("Simu: %s %d", k, len(simuMap[k]))
			go simulateResourcePatrol(k, isVehicle, patrolStopChan)
		}
	}

	missionChanMap = make(map[string]chan string)

	// Simulates battery drop every 5 seconds
	go simulateBatteryDrop(patrolStopChan)
}

func getWayPoints(fStr []byte) [][]float64 {

	var waypoints [][]float64

	scanner := bufio.NewScanner(bytes.NewReader(fStr))
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), ",")
		if len(cols) == 3 && cols[0] != "id" && cols[0] != "squadId" {

			if lat, err := strconv.ParseFloat(strings.TrimSpace(cols[1]), 64); err == nil {
				if lon, err := strconv.ParseFloat(strings.TrimSpace(cols[2]), 64); err == nil {
					waypoints = append(waypoints, []float64{lat, lon})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error(err.Error())
	}
	return waypoints
}

func simulateResourcePatrol(resourceId string, isVehicle bool, stopChan chan int) {
	i := 0
	var stopped = false

	var drone models.DroneH3D
	for i, d := range drones {
		if d.DroneId == resourceId {
			drone = drones[i]
			break
		}
	}

	for {
		select {
		case <-stopChan:
			log.Info("Produce for ID: %s received stopped", resourceId)
			// mission is stopped
			stopped = true
		default:
		}
		if stopped {
			log.Info("Produce for ID: %s stopped", resourceId)
			break
		}
		i = i % len(simuMap[resourceId])
		if resourceStatusMap[resourceId] == patrolStatus {
			res := models.Resource{ID: resourceId,
				Type:      "",
				Name:      "",
				Latitude:  simuMap[resourceId][i][0],
				Longitude: simuMap[resourceId][i][1],
			}
			log.Info("Produce for ID: %s  %d locations: %f %f", res.ID, i, res.Latitude, res.Longitude)
			locChan <- res
			location := strconv.FormatFloat(res.Latitude, 'E', -1, 64) + "," + strconv.FormatFloat(res.Longitude, 'E', -1, 64)
			loc := models.ResourceLocation{
				ResourceId:  resourceId,
				Location:    location,
				Altitude:    drone.CurrAltitude,
				IsExternal:  true,
				IsVehicle:   isVehicle,
				TimestampMs: time.Now().UnixNano() / int64(time.Millisecond),
			}
			sendLocation(loc)
			i++
		}
		// Sleep for 20 seconds
		time.Sleep(time.Second * 20)
	}
}

func startResourceMission(mission models.MissionCommand, isVehicle bool) error {
	currentLat := -1.0
	currentLon := -1.0
	for _, res := range resources {
		if res.ID == *(mission.ResourceId) {
			currentLat = res.Latitude
			currentLon = res.Longitude
		}
	}
	if currentLat < 0 || currentLon < 0 {
		return errors.New("resource cannot be found")
	}
	waypoints := getStraightRoute([]float64{currentLat, currentLon}, mission.Waypoints[0])
	i := 0
	startMission := true
	resourceStatusMap[*mission.ResourceId] = missionStatus
	missionChan, found := missionChanMap[*mission.ResourceId]

	if !found {
		log.Error("%s mission channel not found", *mission.ResourceId)
		return nil
	}

	for {

		// need to check whether the stop mission is initiated
		select {
		case msg := <-missionChan:
			log.Info("%s received stop mission message: %s", *(mission.ResourceId), msg)
			startMission = false
			// mission is stopped
		default:
		}
		// Sleep for 5 seconds
		time.Sleep(time.Second * 5)

		if !startMission {
			log.Info("%s mission completed", *(mission.ResourceId))
			break
		} else if i == (len(waypoints) - 1) {
			log.Info("%s attending to mission %s", *(mission.ResourceId), *mission.MissionId)
			continue
		} else {
			i++
		}

		res := models.Resource{ID: *mission.ResourceId,
			Type:      "",
			Name:      "",
			Latitude:  waypoints[i][0],
			Longitude: waypoints[i][1],
		}
		log.Info("Produce mission for ID: %s  %d locations: %f %f", res.ID, i, res.Latitude, res.Longitude)
		locChan <- res
		location := strconv.FormatFloat(res.Latitude, 'E', -1, 64) + "," + strconv.FormatFloat(res.Longitude, 'E', -1, 64)
		loc := models.ResourceLocation{
			ResourceId:  *mission.ResourceId,
			Location:    location,
			Altitude:    float64(*applicationConfig.Altitude),
			IsExternal:  true,
			IsVehicle:   isVehicle,
			TimestampMs: time.Now().UnixNano() / int64(time.Millisecond),
		}
		sendLocation(loc)
	}
	droneId := *mission.ResourceId
	go goBackToBase(droneId)

	return nil
}

// base on source & dest with straight line and 2 mins movement to return the route
func getStraightRoute(source []float64, dest []float64) [][]float64 {
	// lat & lon
	if len(source) != 2 || len(dest) != 2 {
		return nil
	}
	log.Info("Route: source: %f %f dest: %f %f", source[0], source[1], dest[0], dest[1])
	waypoints := make([][]float64, 13)
	latDiff := (dest[0] - source[0]) / 12
	lonDiff := (dest[1] - source[1]) / 12

	waypoints[0] = make([]float64, 2)
	waypoints[0][0] = source[0]
	waypoints[0][1] = source[1]
	for i := 1; i <= 12; i++ {
		waypoints[i] = make([]float64, 2)
		waypoints[i][0] = waypoints[i-1][0] + latDiff
		waypoints[i][1] = waypoints[i-1][1] + lonDiff
	}
	return waypoints
}

func updateLocations() {
	for {
		loc := <-locChan
		// log.Info("Consume with ID: %s locations: %f %f", loc.ID, loc.Latitude, loc.Longitude)

		for i, res := range resources {
			if res.ID == loc.ID {
				resources[i].Latitude = loc.Latitude
				resources[i].Longitude = loc.Longitude
				break
			}
		}

		for i, drone := range drones {
			if drone.DroneId == loc.ID {
				heading := getHeadingBetweenCoordinates([]float64{drones[i].CurrLat, drones[i].CurrLong}, []float64{loc.Latitude, loc.Longitude})
				drones[i].CurrHeading = heading
				drones[i].CurrLat = loc.Latitude
				drones[i].CurrLong = loc.Longitude
				break
			}
		}

		time.Sleep(1 * time.Millisecond)

	}
}

// func initiateDrone(droneId string) {
// 	// Add to Drone Connector Managed Drones List
// 	getUrl := *applicationConfig.RestAPIAddress + *applicationConfig.ResourcesBasePath + "/" + droneId
// 	log.Info("ManageDrone API Get " + getUrl)

// 	resp, err := httpClient.Get(getUrl)
// 	// Check for errors when doing REST API Post
// 	if err != nil {
// 		log.Error("Managing Drone REST Error: %s", err.Error())
// 	} else if resp != nil {
// 		if resp.Body != nil {
// 			defer resp.Body.Close()
// 		}
// 		if resp.StatusCode != http.StatusOK {
// 			log.Info("Managing Drone REST Error: %d", resp.StatusCode)
// 		}
// 	}
// }

// Calculate distance in km using Pythagoras' theorem
func getDistanceBetweenCoordinates(start []float64, end []float64) float64 {
	phi1 := start[0] * math.Pi / 180.0
	phi2 := end[0] * math.Pi / 180.0
	lambda1 := start[1] * math.Pi / 180.0
	lambda2 := end[1] * math.Pi / 180.0

	R := 6371000 // Radius of Earth in metres

	x := (lambda2 - lambda1) * math.Cos((phi1+phi2)*math.Pi/2)
	y := phi2 - phi1
	distance := math.Sqrt(math.Pow(x, 2)+math.Pow(y, 2)) * float64(R) / 1000.0
	return distance
}

func getHeadingBetweenCoordinates(start []float64, end []float64) float64 {
	phi1 := start[0] * math.Pi / 180.0
	phi2 := end[0] * math.Pi / 180.0
	lambda1 := start[1] * math.Pi / 180.0
	lambda2 := end[1] * math.Pi / 180.0

	y := math.Sin(lambda2-lambda1) * math.Cos(phi2)
	x := math.Cos(phi1)*math.Sin(phi2) - math.Sin(phi1)*math.Cos(phi2)*math.Cos(lambda2-lambda1)
	theta := math.Atan2(y, x)
	heading := math.Mod((theta*180.0/math.Pi)+360.0, 360.0)
	return math.Round(heading)
}

func sendDroneStatus(drone models.DroneH3D) error {
	var h3dDrone = models.DroneH3dStatus{
		BattLevel:        strconv.Itoa(int(math.Round(drone.BattLevel))),
		DistanceFromHome: fmt.Sprintf("%.1f", getDistanceBetweenCoordinates([]float64{drone.HomeLat, drone.HomeLong}, []float64{drone.CurrLat, drone.CurrLong})),
		DronesPosition:   fmt.Sprint(drone.CurrLat) + "," + fmt.Sprint(drone.CurrLong),
		GpsStatus:        randomInt(5, 7),
		HomePosition:     fmt.Sprint(drone.HomeLat) + "," + fmt.Sprint(drone.HomeLong),
		NetworkType:      drone.NetworkType,
		SignalStrength:   drone.SignalStrength,
		Temperature:      drone.Temperature,
		TextualStatus:    drone.TextualStatus,
	}

	if drone.CurrLat == drone.HomeLat && drone.CurrLong == drone.HomeLong {
		h3dDrone.Altitude = 0
		h3dDrone.DroneSpeed = "0 mph"
		h3dDrone.CurrHeading = 0
	} else {
		h3dDrone.Altitude = int(drone.CurrAltitude)
		h3dDrone.DroneSpeed = strconv.Itoa(*applicationConfig.DroneSpeed) + " mph"
		h3dDrone.CurrHeading = int(drone.CurrHeading)
	}
	var droneStatus = models.TransformDroneStatusFromH3dStatus(h3dDrone, drone.DroneId)
	jsonDroneStatus, err := json.Marshal(droneStatus)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("Produce for ID: %s statuses: %+v", drone.DroneId, droneStatus)

	postUrl := *applicationConfig.RestAPIAddress + *applicationConfig.ResourcesBasePath + "/" + drone.DroneId + "/status"
	// log.Info("sendDroneStatus for %s", drone.DroneId)

	resp, err := httpClient.Post(postUrl, "application/json", bytes.NewBuffer(jsonDroneStatus))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	return nil
}

func sendLocation(loc models.ResourceLocation) error {

	postUrl := *applicationConfig.RestAPIAddress + *applicationConfig.ResourcesBasePath + "/" + loc.ResourceId + "/location"
	// log.Info("sendLocations for %s", loc.ResourceId)

	jsonLoc, err := json.Marshal(loc)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	resp, err := httpClient.Post(postUrl, "application/json", bytes.NewBuffer(jsonLoc))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	return nil
}

func simulateBatteryDrop(stopChan chan int) {
	messageInterval := 5
	depletionRate := float64(100) / (*applicationConfig.BatteryLife * 60)
	for {
		select {
		case <-stopChan:
			log.Info("simulateBatteryDrop received stopped")
			// mission is stopped
		default:
		}
		// To keep drones actively managed in Drone Connector
		// To change when doing autodiscovery
		for i := 0; i < len(drones); i++ {
			// Drone is charging at base
			if drones[i].CurrLat == drones[i].HomeLat && drones[i].CurrLong == drones[i].HomeLong {
				batteryLevel := drones[i].BattLevel + 10
				if batteryLevel > 100 {
					drones[i].BattLevel = 100
				} else {
					drones[i].BattLevel = batteryLevel
				}
			} else if drones[i].BattLevel <= 0 {
				// Teleport drone back to base
				res := models.Resource{
					ID:        drones[i].DroneId,
					Type:      "",
					Name:      "",
					Latitude:  drones[i].HomeLat,
					Longitude: drones[i].HomeLong,
				}
				locChan <- res

				isVehicle := false
				for _, res := range resources {
					if res.ID == drones[i].DroneId {
						isVehicle = res.IsVehicle
						break
					}
				}

				location := strconv.FormatFloat(drones[i].HomeLat, 'E', -1, 64) + "," + strconv.FormatFloat(drones[i].HomeLong, 'E', -1, 64)
				loc := models.ResourceLocation{
					ResourceId:  drones[i].DroneId,
					Location:    location,
					Altitude:    0,
					IsExternal:  true,
					IsVehicle:   isVehicle,
					TimestampMs: time.Now().UnixNano() / int64(time.Millisecond),
				}
				sendLocation(loc)
			} else {
				batteryLevel := float64(drones[i].BattLevel) - (float64(messageInterval) * depletionRate)
				if batteryLevel < 0 {
					drones[i].BattLevel = 0
				} else {
					drones[i].BattLevel = batteryLevel
				}
			}
			sendDroneStatus(drones[i])
		}
		// Sleep for 5 seconds
		time.Sleep(time.Second * time.Duration(messageInterval))
	}
}

func GetDroneInfo(rc *models.RequestContext, droneId string) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetDroneInfo from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Requesting Drone's Info." + rc.EchoContext.Request().RequestURI)
	}

	i := checkIndex(droneId)
	if i >= 0 {
		var h3dDrone = models.DroneH3dStatus{
			Altitude:         *applicationConfig.Altitude,
			BattLevel:        strconv.Itoa(int(math.Round(drones[i].BattLevel))),
			DistanceFromHome: fmt.Sprintf("%.1f", getDistanceBetweenCoordinates([]float64{drones[i].HomeLat, drones[i].HomeLong}, []float64{drones[i].CurrLat, drones[i].CurrLong})),
			DroneSpeed:       strconv.Itoa(*applicationConfig.DroneSpeed) + " mph",
			DronesPosition:   fmt.Sprint(drones[i].CurrLat) + "," + fmt.Sprint(drones[i].CurrLong),
			GpsStatus:        randomInt(5, 7),
			CurrHeading:      randomInt(0, 360),
			HomePosition:     fmt.Sprint(drones[i].HomeLat) + "," + fmt.Sprint(drones[i].HomeLong),
			NetworkType:      drones[i].NetworkType,
			SignalStrength:   *applicationConfig.SignalStrength,
			Temperature:      strconv.Itoa(*applicationConfig.Temperature),
			TextualStatus:    drones[i].TextualStatus,
		}
		json_data, err := json.Marshal(h3dDrone)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// Printing out Mission Details using regex
		// log.Info("Synchronous Response -> Drone Details:")
		// log.Info("%#v", h3dDrone)

		if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
			rc.EchoContext.Response().Header().Set("Content-Type", "application/json")
			rc.EchoContext.Response().WriteHeader(http.StatusOK)
			rc.EchoContext.Response().Writer.Write(json_data)
		}
	}
	return nil
}

func GetDroneVideo(rc *models.RequestContext, droneId string) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetDroneVideo from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Requesting Drone's Video." + rc.EchoContext.Request().RequestURI)
	}

	index := checkIndex(droneId)
	if index >= 0 {
		json_data, err := json.Marshal(drones[index].DroneVideo)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// Printing out Mission Details using regex
		log.Info("Synchronous Response -> Drone Details:")
		log.Info("%#v", drones[index].DroneVideo)

		if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
			rc.EchoContext.Response().Header().Set("Content-Type", "application/json")
			rc.EchoContext.Response().WriteHeader(http.StatusOK)
			rc.EchoContext.Response().Writer.Write(json_data)
		}
	}
	return nil
}

func GetAllDrones(rc *models.RequestContext) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetAllDrones from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Requesting All Drone's Info." + rc.EchoContext.Request().RequestURI)
	}

	dronesH3d := models.TransformDroneH3dFromDrone(drones)
	json_data, err := json.Marshal(dronesH3d)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Printing out Mission Details using regex
	log.Info("Synchronous Response -> All Drone Details:")
	log.Info("%#v", dronesH3d)

	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
		rc.EchoContext.Response().Header().Set("Content-Type", "application/json")
		rc.EchoContext.Response().WriteHeader(http.StatusOK)
		rc.EchoContext.Response().Writer.Write(json_data)
	}
	return nil
}

func GetAllResources(rc *models.RequestContext) ([]models.Resource, error) {
	return resources, nil
}

func GetAllFlights(rc *models.RequestContext) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetAllFlights from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Requesting All Drone Flights." + rc.EchoContext.Request().RequestURI)
	}

	json_data, err := json.Marshal(drones)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Printing out Mission Details using regex
	log.Info("Synchronous Response -> All Drone Details:")
	log.Info("%#v", drones)

	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
		rc.EchoContext.Response().Header().Set("Content-Type", "application/json")
		rc.EchoContext.Response().WriteHeader(http.StatusOK)
		rc.EchoContext.Response().Writer.Write(json_data)
	}
	return nil
}

func GetAllDroneServers(rc *models.RequestContext) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetAllDroneServers from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Requesting All Drone Servers." + rc.EchoContext.Request().RequestURI)
	}

	json_data, err := json.Marshal(drones)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Printing out Mission Details using regex
	log.Info("Synchronous Response -> All Drone Details:")
	log.Info("%#v", drones)

	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
		rc.EchoContext.Response().Header().Set("Content-Type", "application/json")
		rc.EchoContext.Response().WriteHeader(http.StatusOK)
		rc.EchoContext.Response().Writer.Write(json_data)
	}
	return nil
}

/*func simulateStartStopMission(c echo.Context) error {
	log.Info("Received REST Start or Stop Mission Command from", r.Host, r.URL.Path)
	w.WriteHeader(http.StatusOK)
}*/

func checkIndex(droneId string) int {
	index := -1
	for i := 0; i < len(drones); i++ {
		if droneId == drones[i].DroneId {
			index = i
		}
	}
	return index
}

func StartResourceMission(mission models.MissionCommand) error {
	var missionChan chan string
	var exists = false
	var isVehicle = false
	for _, res := range resources {
		if res.ID == *mission.ResourceId {
			exists = true
			isVehicle = res.IsVehicle
		}
	}
	if !exists {
		return errors.New("resource not found")
	}

	// Create resource channel if not exists
	if theChan, found := missionChanMap[*mission.ResourceId]; found {
		missionChan = theChan
	} else {
		missionChan = make(chan string)
		missionChanMap[*mission.ResourceId] = missionChan
	}

	// Check current status of resource
	if resourceStatusMap[*mission.ResourceId] == returnToBaseStatus {
		go func(messageChan chan string) {
			messageChan <- "START"
			fmt.Println("sent message", "START")
			startResourceMission(mission, isVehicle)
		}(missionChan)
	} else if resourceStatusMap[*mission.ResourceId] == missionStatus {
		return errors.New(*mission.ResourceId + " is not available")
	} else {
		go startResourceMission(mission, isVehicle)
	}

	return nil
}

func StopResourceMission(mission models.MissionCommand) error {
	var exists = false
	for _, res := range resources {
		if res.ID == *mission.ResourceId {
			exists = true
		}
	}
	if !exists {
		return errors.New("resource not found")
	}
	if theChan, found := missionChanMap[*mission.ResourceId]; found {
		go func(messageChan chan string) {
			messageChan <- "STOP"
			fmt.Println("sent message", "STOP")
		}(theChan)
	}
	return nil
}

func Dispose() {

}

// func StartMission(rc *models.RequestContext, missionDetails models.Mission) error {
// 	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
// 		log.Info("Received REST StartMission from " + rc.EchoContext.Request().RemoteAddr)
// 		log.Info("Start Mission Command." + rc.EchoContext.Request().RequestURI)
// 	}

// 	// Check for Valid Mission Details
// 	if checkValidMissionDetails(missionDetails) {
// 		index := checkIndex(*missionDetails.DroneId)
// 		if index >= 0 {
// 			drones[index].Mission = missionDetails
// 			go startMissionNow(&drones[index], index)
// 		}
// 	} else {
// 		rc.EchoContext.Response().Writer.Write([]byte(`{"message": "Invalid Mission Details (Failed Check)","status": "failure"}`))
// 		log.Info("Error in startMission Command, Invalid Mission Details")
// 		return fmt.Errorf("invalid Mission details")
// 	}

// 	// Printing out Mission Details using regex
// 	log.Info("Mission Details:")
// 	log.Info("%#v", missionDetails)

// 	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
// 		rc.EchoContext.Response().WriteHeader(http.StatusOK)
// 		rc.EchoContext.Response().Writer.Write([]byte(`{"message": "Successfully Called startMission Command over REST","status": "success"}`))
// 	}
// 	return nil
// }

// func StopMission(rc *models.RequestContext, droneId string) error {
// 	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
// 		log.Info("Received REST StopMission from " + rc.EchoContext.Request().RemoteAddr)
// 		log.Info("Start Mission Command." + rc.EchoContext.Request().RequestURI)
// 	}

// 	index := checkIndex(droneId)
// 	if index >= 0 {
// 		drones[index].Mission.Waypoints = [][]float64{{drones[index].HomeLong, drones[index].HomeLat}}
// 		go startMissionNow(&drones[index], index)
// 	}

// 	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
// 		rc.EchoContext.Response().WriteHeader(http.StatusOK)
// 		rc.EchoContext.Response().Writer.Write([]byte(`{"message": "Successfully Called stopMission Command over REST","status": "success"}`))
// 	}
// 	return nil
// }

func GetMissionDetails(rc *models.RequestContext, droneId string) error {
	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Request() != nil {
		log.Info("Received REST GetMissionDetails from " + rc.EchoContext.Request().RemoteAddr)
		log.Info("Get Mission Details." + rc.EchoContext.Request().RequestURI)
	}
	missionDetails := models.Mission{
		//MissionId:   drone1.Mission.MissionId,
		NewMission: models.NewMission{
			MissionName: drones[0].Mission.NewMission.MissionName,
		},
		Waypoints: drones[0].Mission.Waypoints,
	}
	json_data, err := json.Marshal(missionDetails)

	// Printing out Mission Details using regex
	log.Info("Synchronous Response -> Mission Details:")
	log.Info("%#v", missionDetails)

	if rc != nil && rc.EchoContext != nil && rc.EchoContext.Response() != nil {
		// Check for errors during Marshalling of JSON
		if err != nil {
			rc.EchoContext.Response().WriteHeader(http.StatusInternalServerError)
			rc.EchoContext.Response().Writer.Write([]byte(`{"message": "Invalid Mission Details (Marshalling Error)!","status": "failure","cmdId" : 2}`))
			log.Error(err.Error())
			return err
		}
		rc.EchoContext.Response().WriteHeader(http.StatusOK)
		rc.EchoContext.Response().Writer.Write(json_data)
	}
	return nil
}

// func checkValidMissionDetails(missionDetails models.Mission) bool {
// 	if len(missionDetails.Waypoints) == 0 {
// 		return false
// 	}
// 	for i := 0; i < len(missionDetails.Waypoints); i++ {
// 		if len(missionDetails.Waypoints[i]) != 2 {
// 			return false
// 		}
// 	}
// 	return true
// }

// Simulate Movement for Drone
// func simulateMovement() {
// 	for i := 0; i < len(drones); i++ {
// 		drones[i].Mission.Waypoints = [][]float64{{float64(103.835984), float64(1.342184)}, {float64(103.835984), float64(1.422184)}, {float64(103.755984), float64(1.402184)}, {float64(103.785984), float64(1.352184)}, {float64(103.835984), float64(1.342184)}}
// 		go startMissionNow(&drones[i], i)
// 	}
// }

// Using simulated Lat Long movements for now
// func startMissionNow(drone *models.DroneH3D, droneIndex int) {
// 	log.Info("............................................... Starting Mission Now ...............................................")
// 	if checkIndex(drone.DroneId) >= 0 {
// 		missionStatuses[checkIndex(drone.DroneId)] = 1
// 		currWayPoint := 0
// 		halfPi := float64(0.5 * math.Pi)
// 		var inRads float64

// 		for currWayPoint < len(drone.Mission.Waypoints) {
// 			droneCurrLong := drone.CurrLong
// 			droneCurrLat := drone.CurrLat
// 			log.Info("->->->->->->->->->->-> Going to Waypoint %d Longitude: %f, Latitude: %f <-<-<-<-<-<-<-<-<-<-<-", currWayPoint+1, drone.Mission.Waypoints[currWayPoint][0], drone.Mission.Waypoints[currWayPoint][1])
// 			var latIncrement, longIncrement float64
// 			if drone.Mission.Waypoints[currWayPoint][0] == droneCurrLong {
// 				longIncrement = 0
// 				if drone.Mission.Waypoints[currWayPoint][1] > droneCurrLat {
// 					inRads = halfPi
// 					latIncrement = *applicationConfig.DroneSpeed
// 				} else {
// 					inRads = -halfPi
// 					latIncrement = -*applicationConfig.DroneSpeed
// 				}
// 			} else if drone.Mission.Waypoints[currWayPoint][1] == droneCurrLat {
// 				latIncrement = 0
// 				if drone.Mission.Waypoints[currWayPoint][0] > droneCurrLong {
// 					inRads = 0
// 					longIncrement = *applicationConfig.DroneSpeed
// 				} else {
// 					inRads = -math.Pi
// 					longIncrement = -*applicationConfig.DroneSpeed
// 				}
// 			} else {
// 				lon1 := droneCurrLong * math.Pi / 180.0
// 				lon2 := drone.Mission.Waypoints[currWayPoint][0] * math.Pi / 180.0
// 				dLon := lon2 - lon1
// 				lat1 := droneCurrLat * math.Pi / 180.0
// 				lat2 := drone.Mission.Waypoints[currWayPoint][1] * math.Pi / 180.0
// 				x := math.Sin(dLon) * math.Cos(lat2)
// 				y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
// 				inRads = float64(math.Atan2(y, x))
// 				log.Info("lon1:%f, lon2:%f, lat1:%f, lat2:%f, y:%f, x:%f, Radians :%f", lon1, lon2, lat1, lat2, y, x, inRads)

// 				if inRads > halfPi {
// 					ratio := math.Tan(inRads - halfPi)
// 					latIncrement = 1 / (ratio + 1) * *applicationConfig.DroneSpeed
// 					longIncrement = -ratio / (ratio + 1) * *applicationConfig.DroneSpeed
// 					log.Info("> halfPi, ratio:%f, Long Increment:%f, Lat Increment:%f", ratio, longIncrement, latIncrement)
// 				} else if inRads >= 0 {
// 					ratio := math.Tan(halfPi - inRads)
// 					latIncrement = 1 / (ratio + 1) * *applicationConfig.DroneSpeed
// 					longIncrement = ratio / (ratio + 1) * *applicationConfig.DroneSpeed
// 					log.Info("> 0, < halfPi, ratio:%f, Long Increment:%f, Lat Increment:%f", ratio, longIncrement, latIncrement)
// 				} else if inRads < -halfPi {
// 					ratio := math.Tan(-inRads - halfPi)
// 					latIncrement = -1 / (ratio + 1) * *applicationConfig.DroneSpeed
// 					longIncrement = -ratio / (ratio + 1) * *applicationConfig.DroneSpeed
// 					log.Info("< -halfPi, ratio:%f, Long Increment:%f, Lat Increment:%f", ratio, longIncrement, latIncrement)
// 				} else { // inRads < 0, but > -0.5*math.Pi
// 					ratio := math.Tan(halfPi + inRads)
// 					latIncrement = -1 / (ratio + 1) * *applicationConfig.DroneSpeed
// 					longIncrement = ratio / (ratio + 1) * *applicationConfig.DroneSpeed
// 					log.Info("< 0, > -halfPi, ratio:%f, Long Increment:%f, Lat Increment:%f", ratio, longIncrement, latIncrement)
// 				}
// 			}

// 			heading := float64(inRads)
// 			drone.CurrHeading = heading
// 			log.Info("Current Heading:%f", drone.CurrHeading)

// 			if inRads > halfPi {
// 				for droneCurrLat < drone.Mission.Waypoints[currWayPoint][1] || droneCurrLong > drone.Mission.Waypoints[currWayPoint][0] {
// 					incrementLatLong(droneCurrLat, droneCurrLong, drone.Mission.Waypoints[currWayPoint][1], drone.Mission.Waypoints[currWayPoint][0], latIncrement, longIncrement, droneIndex)
// 					droneCurrLong = drone.CurrLong
// 					droneCurrLat = drone.CurrLat
// 				}
// 			} else if inRads >= 0 {
// 				for droneCurrLat < drone.Mission.Waypoints[currWayPoint][1] || droneCurrLong < drone.Mission.Waypoints[currWayPoint][0] {
// 					incrementLatLong(droneCurrLat, droneCurrLong, drone.Mission.Waypoints[currWayPoint][1], drone.Mission.Waypoints[currWayPoint][0], latIncrement, longIncrement, droneIndex)
// 					droneCurrLong = drone.CurrLong
// 					droneCurrLat = drone.CurrLat
// 				}
// 			} else if inRads < -halfPi {
// 				for droneCurrLat > drone.Mission.Waypoints[currWayPoint][1] || droneCurrLong > drone.Mission.Waypoints[currWayPoint][0] {
// 					incrementLatLong(droneCurrLat, droneCurrLong, drone.Mission.Waypoints[currWayPoint][1], drone.Mission.Waypoints[currWayPoint][0], latIncrement, longIncrement, droneIndex)
// 					droneCurrLong = drone.CurrLong
// 					droneCurrLat = drone.CurrLat
// 				}
// 			} else { // inRads < 0, but > -0.5*math.Pi
// 				for droneCurrLat > drone.Mission.Waypoints[currWayPoint][1] || droneCurrLong < drone.Mission.Waypoints[currWayPoint][0] {
// 					incrementLatLong(droneCurrLat, droneCurrLong, drone.Mission.Waypoints[currWayPoint][1], drone.Mission.Waypoints[currWayPoint][0], latIncrement, longIncrement, droneIndex)
// 					droneCurrLong = drone.CurrLong
// 					droneCurrLat = drone.CurrLat
// 				}
// 			}
// 			currWayPoint++
// 		}
// 		log.Info("............................................... Mission Completed ...............................................")
// 		missionStatuses[checkIndex(drone.DroneId)] = 0
// 	}
// }

// func incrementLatLong(currLat float64, currLong float64, waypointLat float64, waypointLong float64, latIncrement float64, longIncrement float64, droneIndex int) {
// 	if droneIndex >= 0 {
// 		if math.Abs(currLat-waypointLat) <= math.Abs(latIncrement) {
// 			currLat = waypointLat
// 		} else {
// 			currLat += latIncrement
// 		}
// 		if math.Abs(currLong-waypointLong) <= math.Abs(longIncrement) {
// 			currLong = waypointLong
// 		} else {
// 			currLong += longIncrement
// 		}

// 		drones[droneIndex].CurrLat = currLat
// 		drones[droneIndex].CurrLong = currLong

// 		// sleep for 1 second
// 		time.Sleep(time.Second)
// 	}
// }

func chunkResources(resources []models.Resource) [][]models.Resource {
	chunks := make([][]models.Resource, 0)
	for i := 0; i < len(resources); i += 500 {
		last := i + 500
		if last > len(resources) {
			last = len(resources)
		}
		chunks = append(chunks, resources[i:last])
	}
	return chunks
}

func getToken() (string, error) {
	postUrl := *applicationConfig.KeycloakTokenUrl

	payload := strings.NewReader("grant_type=password&scope=openid%20profile%20email&username=corentin.dodon&password=1234567&client_id=sdpuiauth")

	resp, err := httpClient.Post(postUrl, "application/x-www-form-urlencoded", payload)

	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	var j struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return j.AccessToken, nil
}

func getResource(droneId string, token string) (string, error) {

	getUrl := *applicationConfig.RmsRestAPIAddress + *applicationConfig.RmsResourceBasePath + "/" + droneId

	newRequest, newRequestError := http.NewRequest("GET", getUrl, nil)
	if newRequestError != nil {
		log.Error(newRequestError.Error())
		return "", newRequestError
	}
	newRequest.Header.Add("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(newRequest)

	// Check for errors when doing REST API Get
	if err != nil {
		log.Error("Managing Drone REST Error: %s", err.Error())
		return "", err

	} else if resp != nil {
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if resp.StatusCode != http.StatusOK {
			log.Info("Managing Drone REST Error: %d", resp.StatusCode)
			return "", errors.New("HTTP status: " + strconv.Itoa(resp.StatusCode))
		}
	}

	var j struct {
		SquadId string `json:"squadId"`
	}
	jsonErr := json.NewDecoder(resp.Body).Decode(&j)
	if jsonErr != nil {
		return "", jsonErr
	}

	return j.SquadId, nil
}

func updateDroneSquadOperationalStatus(droneId string, status string) {
	token, err := getToken()
	if err != nil {
		return
	}

	squadId, err := getResource(droneId, token)
	if err != nil {
		return
	}

	putUrl := *applicationConfig.RmsRestAPIAddress + *applicationConfig.RmsSquadsBasePath + "/" + squadId

	values := map[string]string{"status": status}
	jsonValue, _ := json.Marshal(values)

	newRequest, newRequestError := http.NewRequest("PUT", putUrl, bytes.NewBuffer(jsonValue))
	if newRequestError != nil {
		log.Error(newRequestError.Error())
		return
	}
	newRequest.Header.Add("Authorization", "Bearer "+token)

	httpClient.Do(newRequest)
}

func goBackToBase(resourceId string) error {
	resourceStatusMap[resourceId] = returnToBaseStatus

	currentLat := -1.0
	currentLon := -1.0
	baseLat := -1.0
	baseLon := -1.0

	for _, res := range resources {
		if res.ID == resourceId {
			currentLat = res.Latitude
			currentLon = res.Longitude
			baseLat = res.BaseLatitude
			baseLon = res.BaseLongitude
		}
	}
	if currentLat < 0 || currentLon < 0 {
		return errors.New("resource cannot be found")
	}
	waypoints := getStraightRoute([]float64{currentLat, currentLon}, []float64{baseLat, baseLon})
	i := 0
	missionChan := missionChanMap[resourceId]

	for {
		// Listen for new start mission message
		select {
		case msg := <-missionChan:
			log.Info("%s received start mission message: %s, terminating return to base", resourceId, msg)
			resourceStatusMap[resourceId] = patrolStatus
			return nil
		default:
		}

		// Sleep for 5 seconds
		time.Sleep(time.Second * 5)

		// reach the dest, stop update the locations
		if i == (len(waypoints) - 1) {
			log.Info("%s has reached base", resourceId)
			break
		} else {
			i++
		}

		res := models.Resource{ID: resourceId,
			Type:      "",
			Name:      "",
			Latitude:  waypoints[i][0],
			Longitude: waypoints[i][1],
		}
		log.Info("Back to base ID: %s  %d locations: %f %f", res.ID, i, res.Latitude, res.Longitude)
		locChan <- res
		location := strconv.FormatFloat(res.Latitude, 'E', -1, 64) + "," + strconv.FormatFloat(res.Longitude, 'E', -1, 64)
		loc := models.ResourceLocation{
			ResourceId:  resourceId,
			Location:    location,
			Altitude:    float64(*applicationConfig.Altitude),
			IsExternal:  true,
			IsVehicle:   false,
			TimestampMs: time.Now().UnixNano() / int64(time.Millisecond),
		}
		sendLocation(loc)
	}

	resourceStatusMap[resourceId] = patrolStatus

	for i, drone := range drones {
		if drone.DroneId == resourceId {
			drones[i].CurrAltitude = 0
		}
	}

	return nil
}

func getRestrictedZone() []restrictedZone.RestrictedZone {
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
		log.Error("Failed to get response: %v", err)
	}

	var result models.ApiResponse
	if err := json.Unmarshal(response, &result); err != nil {
		log.Error("Failed to unmarshal response: %v", err)
	}
	var restrictedZones []restrictedZone.RestrictedZone
	for _, instance := range result.CeInstances {
		var polygon []restrictedZone.Point
		id := instance.ID
		if outer, ok := instance.Geometry.Coordinates.([]interface{}); ok {
			if middle, ok := outer[0].([]interface{}); ok {
				if inner, ok := middle[0].([]interface{}); ok {
					for _, coordinate := range inner {
						if firstCoordinate, ok := coordinate.([]interface{}); ok {
							// Extract latitude and longitude from the first coordinate
							if len(firstCoordinate) >= 2 {
								longitude := firstCoordinate[0].(float64)
								latitude := firstCoordinate[1].(float64)
								polygon = append(polygon, restrictedZone.Point{Lat: latitude, Lon: longitude})
							}
						}
					}
				}
			}
		}

		restrictedZone := restrictedZone.RestrictedZone{
			ID:        id,
			Polygon:   polygon,
			StartTime: instance.Data.ActivationStart.TimestampMs,
			EndTime:   instance.Data.ActivationEnd.TimestampMs,
		}
		restrictedZones = append(restrictedZones, restrictedZone)
	}
	return restrictedZones
}

func GetSourceDestinationPoints(coordinates string) (restrictedZone.Point, restrictedZone.Point, error) {
	parts := strings.Split(coordinates, ":")

	// Split the first part to get the source lat/lon
	sourceCoords := strings.Split(parts[0], ",")
	sourceLat, _ := strconv.ParseFloat(sourceCoords[0], 64)
	sourceLon, _ := strconv.ParseFloat(sourceCoords[1], 64)
	source = restrictedZone.Point{Lat: sourceLat, Lon: sourceLon}

	// Split the second part to get the destination lat/lon
	destinationCoords := strings.Split(parts[1], ",")
	destinationLat, _ := strconv.ParseFloat(destinationCoords[0], 64)
	destinationLon, _ := strconv.ParseFloat(destinationCoords[1], 64)
	destination = restrictedZone.Point{Lat: destinationLat, Lon: destinationLon}

	fmt.Println(source.Lat)
	fmt.Println(destination.Lat)
	return source, destination, nil
}

func GetPath(rc *models.RequestContext, query string) (float64, []models.ClearanceZone, error) {
	restrictedZones := getRestrictedZone()
	source, destination, nil := GetSourceDestinationPoints(query)

	// Check if the path intersects any active no-fly zones
	currentTime := time.Now()
	droneSpeed := *applicationConfig.DroneSpeed

	isPathInRestrictedZone, crossedZones := restrictedZone.IsPathInRestrictedZone(source, destination, restrictedZones, currentTime, float64(droneSpeed))
	distance := haversineDistance(source, destination)
	distanceInMeters := distance * 1000
	if isPathInRestrictedZone {
		fmt.Println("The path intersects an active no-fly zone!")
		return distanceInMeters, crossedZones, nil
	} else {
		fmt.Println("The path does not intersect any active no-fly zone.")
		return distanceInMeters, crossedZones, nil
	}

}

func GetRemainingOperationTimeAtLocation(resourceId string, travelTimeInSeconds float64) float64 {
	i := checkIndex(resourceId)
	depletionRate := float64(100) / (*applicationConfig.BatteryLife * 60)
	batteryLevel := float64(drones[i].BattLevel) - (travelTimeInSeconds * depletionRate)
	remainingOperationTimeAtLocation := (1 / depletionRate) * batteryLevel
	return remainingOperationTimeAtLocation
}

func haversineDistance(p1, p2 restrictedZone.Point) float64 {
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

func interpolate(p1, p2 restrictedZone.Point, fraction float64) models.Point {
	return models.Point{
		Latitude:  p1.Lat + (p2.Lat-p1.Lat)*fraction,
		Longitude: p1.Lon + (p2.Lon-p1.Lon)*fraction,
	}
}

func GetWaypoints(source, destination restrictedZone.Point, startTime, endTime time.Time, numWaypoints int) []models.Point {
	// Calculate total distance and total time
	//totalDistance := haversine(source, destination)
	totalTime := endTime.Sub(startTime)

	// Calculate time per waypoint
	timeStep := totalTime / time.Duration(numWaypoints-1)

	// Generate waypoints with timestamps
	waypoints := make([]models.Point, numWaypoints)

	for i := 0; i < numWaypoints; i++ {
		fraction := float64(i) / float64(numWaypoints-1) // Progress fraction [0, 1]
		waypoint := interpolate(source, destination, fraction)
		waypoint.Time = startTime.Add(timeStep * time.Duration(i))
		waypoints[i] = waypoint
	}

	return waypoints
}
