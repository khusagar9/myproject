package controller

import (
	"fmt"
	"h3d-drone-emulator/config"
	"h3d-drone-emulator/models"
	"h3d-drone-emulator/service"
	"net/http"
	"strings"

	"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/log"

	restrictedZone "h3d-drone-emulator/util"
	"time"

	"github.com/labstack/echo/v4"
)

type Emulator struct {
}

// NewDroneSubsystem Constructor
func NewEmulatorSubsystem() *Emulator {
	co := new(Emulator)
	return co
}

var appConfig config.AppConfig

/*
* Public functions
 */
// InitDroneConnector Initialize Controller
func (co *Emulator) Initialize(e *echo.Echo) {
	appConfig = config.Get()
	pathParamDroneId := "/:drone_id"
	// REST Classic APIs
	groupRest := e.Group(*appConfig.EndPointUrl + *appConfig.VersionPath)
	groupRest.GET(*appConfig.GetHealthPath, co.isHealthy)
	groupRest.GET(*appConfig.DroneBasePath+pathParamDroneId+*appConfig.DroneInfoPath, co.getDroneInfo)
	groupRest.GET(*appConfig.DroneBasePath+pathParamDroneId+*appConfig.DroneVideoPath, co.getDroneVideo)
	groupRest.GET(*appConfig.AllDronesPath, co.getAllDrones)
	groupRest.GET(*appConfig.AllDroneServersPath, co.getAllDroneServers)
	// groupRest.GET(*appConfig.DroneBasePath+pathParamDroneId+*appConfig.StopMissionPath, co.stopMission)
	// groupRest.POST(*appConfig.StartMissionPath, co.startMission)
	groupRest.POST(*appConfig.AllFlightsPath, co.getAllFlights)
	groupRest.POST(*appConfig.GetMissionPath, co.getMissionDetails)
	groupRest.GET("/resources", co.getAllResources)
	groupRest.POST("/mission/start", co.startResourceMission)
	groupRest.POST("/mission/stop", co.StopResourceMission)
	groupRest.GET(*appConfig.GetRoutePath, co.getRouteDetails)
}

// isHealthy godoc
// @Summary Check for Service health
// @Description Checks for Service health.
// @ID healthy-is
// @Tags EmulatorController
// @Accept  json
// @Produce  json
// @Success 200 "Service is healthy"
// @Router /h3d-drone-emulator/v0/health [get]
func (co *Emulator) isHealthy(c echo.Context) error {
	// TODO Manage health service properly
	return c.JSON(http.StatusOK, "Service is healthy")
}

func (co *Emulator) getDroneInfo(c echo.Context) error {
	//log.Info("getDroneInfo")
	rc := models.CreateRequestContext(c)

	droneId, bindErr := bindDroneIdParam(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	log.Debug("DroneId - %s", droneId)
	err := service.GetDroneInfo(rc, droneId)

	if err != nil {
		return handleErrors(c, "getDroneInfo", err)
	}

	return c.JSON(http.StatusOK, droneId)
}

func (co *Emulator) getDroneVideo(c echo.Context) error {
	//log.Info("getDroneVideo")
	rc := models.CreateRequestContext(c)

	droneId, bindErr := bindDroneIdParam(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	log.Debug("DroneId - %s", droneId)
	err := service.GetDroneVideo(rc, droneId)

	if err != nil {
		return handleErrors(c, "getDroneVideo", err)
	}

	return c.JSON(http.StatusOK, droneId)
}

func (co *Emulator) getAllDrones(c echo.Context) error {
	//log.Info("getAllDrones")
	rc := models.CreateRequestContext(c)
	err := service.GetAllDrones(rc)
	if err != nil {
		return handleErrors(c, "getAllDrones", err)
	}
	return c.JSON(http.StatusOK, "Getting All Drones")
}

func (co *Emulator) getAllResources(c echo.Context) error {

	rc := models.CreateRequestContext(c)
	resources, err := service.GetAllResources(rc)
	if err != nil {
		return handleErrors(c, "getAllResources", err)
	}
	return c.JSON(http.StatusOK, resources)
}

func (co *Emulator) getAllFlights(c echo.Context) error {
	//log.Info("getAllFlights")
	rc := models.CreateRequestContext(c)
	err := service.GetAllFlights(rc)
	if err != nil {
		return handleErrors(c, "getAllFlights", err)
	}
	return c.JSON(http.StatusOK, "Getting All Flights")
}

func (co *Emulator) getAllDroneServers(c echo.Context) error {
	//log.Info("getAllDroneServers")
	rc := models.CreateRequestContext(c)
	err := service.GetAllDroneServers(rc)
	if err != nil {
		return handleErrors(c, "getAllDroneServers", err)
	}
	return c.JSON(http.StatusOK, "Getting All Drone Servers")
}

// func (co *Emulator) startMission(c echo.Context) error {
// 	//log.Info("startMission")
// 	rc := models.CreateRequestContext(c)

// 	mission, bindErr := bindMissionParam(c)
// 	if bindErr != nil {
// 		return handleBadRequest(c, bindErr)
// 	}

// 	log.Debug("Mission - %#v", mission)
// 	err := service.StartMission(rc, *mission)

// 	if err != nil {
// 		return handleErrors(c, "startMission", err)
// 	}

// 	return c.JSON(http.StatusOK, mission)
// }

func (co *Emulator) startResourceMission(c echo.Context) error {

	mission, bindErr := bindMissionCommandParam(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	log.Debug("Resource Mission - %#v", mission)
	err := service.StartResourceMission(*mission)

	if err != nil {
		return handleErrors(c, "startResourceMission", err)
	}

	return c.JSON(http.StatusOK, mission)
}

func (co *Emulator) StopResourceMission(c echo.Context) error {
	mission, bindErr := bindMissionCommandParam(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	log.Debug("Resource Mission - %#v", mission)
	err := service.StopResourceMission(*mission)

	if err != nil {
		return handleErrors(c, "stopResourceMission", err)
	}

	return c.JSON(http.StatusOK, mission)
}

// func (co *Emulator) stopMission(c echo.Context) error {
// 	//log.Info("stopMission")
// 	rc := models.CreateRequestContext(c)

// 	droneId, bindErr := bindDroneIdParam(c)
// 	if bindErr != nil {
// 		return handleBadRequest(c, bindErr)
// 	}

// 	log.Debug("droneId - " + droneId)
// 	err := service.StopMission(rc, droneId)

// 	if err != nil {
// 		return handleErrors(c, "stopMission", err)
// 	}

// 	return c.JSON(http.StatusOK, droneId)
// }

func (co *Emulator) getMissionDetails(c echo.Context) error {
	//log.Info("getMissionDetails")
	rc := models.CreateRequestContext(c)

	droneId, bindErr := bindDroneIdParam(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	err := service.GetMissionDetails(rc, droneId)

	if err != nil {
		return handleErrors(c, "getMissionDetails", err)
	}

	return c.JSON(http.StatusOK, "getMissionDetails")
}

/*func checkValidMissionDetails(missionDetails models.Mission) bool {
	if len(missionDetails.Waypoints) == 0 {
		return false
	}
	for i := 0; i < len(missionDetails.Waypoints); i++ {
		if len(missionDetails.Waypoints[i]) != 2 {
			return false
		}
	}
	return true
}

func bindDroneH3dParam(c echo.Context) (*models.DroneH3D, error) {
	drone := new(models.DroneH3D)
	if err := c.Bind(drone); err != nil {
		log.Error(err.Error())
		return drone, err
	}
	return drone, nil
}

func bindDroneStatusParam(c echo.Context) (*models.DroneH3dStatus, error) {
	droneStatus := new(models.DroneH3dStatus)
	if err := c.Bind(droneStatus); err != nil {
		log.Error(err.Error())
		return droneStatus, err
	}
	return droneStatus, nil
}*/

func bindMissionParam(c echo.Context) (*models.Mission, error) {
	mission := new(models.Mission)
	if err := c.Bind(mission); err != nil {
		log.Error(err.Error())
		return mission, err
	}
	return mission, nil
}

func bindMissionCommandParam(c echo.Context) (*models.MissionCommand, error) {
	mc := new(models.MissionCommand)
	if err := c.Bind(mc); err != nil {

		log.Error(err.Error())
		return nil, err
	}
	return mc, nil
}

func bindDroneIdParam(c echo.Context) (string, error) {
	droneId := c.Param("drone_id")
	if droneId != "" {
		return c.Param("drone_id"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

// Route Path start

func (co *Emulator) getRouteDetails(c echo.Context) error {
	var clearanceRequired bool
	rc := models.CreateRequestContext(c)

	query, bindErr := bindParamQuery(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	routeType, bindErr := bindParamRouteType(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	resorurceId, bindErr := bindParamResourceId(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	travelMode, bindErr := bindParamTravelMode(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	computeBestOrder, bindErr := bindParamComputeBestOrder(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	routeRepresentation, bindErr := bindParamRouteRepresentation(c)
	if bindErr != nil {
		return handleBadRequest(c, bindErr)
	}

	distance, clearenceZonesCrossed, err := service.GetPath(rc, query)

	if err != nil {
		return handleErrors(c, "getRouteDetails", err)

	} else {

		travelTimeInSeconds := calculateTime(distance)

		dispatchTime := *appConfig.DispatchTime
		clearanceTime := *appConfig.ClearanceTime

		clearanceRequired = false

		if len(clearenceZonesCrossed) > 0 {
			// Add clearanceTime and dispatchTime to travelTimeInSeconds
			travelTimeInSeconds = travelTimeInSeconds + float64(clearanceTime+dispatchTime)
			clearanceRequired = true
		} else {
			travelTimeInSeconds = travelTimeInSeconds + float64(dispatchTime)
			clearanceRequired = false
		}

		startTime := time.Now()

		// Format the current time in the desired format
		formattedStartTime := startTime.Format("2006-01-02T15:04:05-07:00")
		fmt.Println("Current Time:", formattedStartTime)

		// Define the number of seconds to add
		secondsToAdd := travelTimeInSeconds // Example: add 120 seconds (2 minutes)

		// Calculate the end time by adding seconds
		endTime := startTime.Add(time.Duration(secondsToAdd) * time.Second)

		// Format the end time in the desired format
		formattedEndTime := endTime.Format("2006-01-02T15:04:05-07:00")
		fmt.Println("End Time after adding seconds:", formattedEndTime)

		numWaypoints := 10

		source, destination, error := service.GetSourceDestinationPoints(query)

		waypoints := service.GetWaypoints(source, destination, startTime, endTime, numWaypoints)

		if error != nil {
			return handleErrors(c, "GetSourceDestinationPoints", err)
		}

		remainingOperationTimeAtLocation := service.GetRemainingOperationTimeAtLocation(resorurceId, travelTimeInSeconds)

		return getRouteRestrictedZone(c, startTime, endTime, clearanceRequired, travelTimeInSeconds, distance, source, destination, waypoints, remainingOperationTimeAtLocation, clearenceZonesCrossed)

	}

	return c.JSON(http.StatusOK, "getRouteDetails"+"-"+query+routeRepresentation+routeType+resorurceId+travelMode+computeBestOrder)
}

func getRouteRestrictedZone(c echo.Context, startTime time.Time, endTime time.Time, clearanceRequired bool, travelTimeInSeconds float64, distance float64, source restrictedZone.Point, destination restrictedZone.Point, waypoints []models.Point, remainingOperationTimeAtLocation float64, clearenceZonesCrossed []models.ClearanceZone) error {
	// Create a mock response
	year, month, day := startTime.Date()
	hour, minute, second := startTime.Clock()

	// Convert to time.Date using the extracted components
	endYear, endMonth, endDay := endTime.Date()
	endHour, endMinute, endSecond := endTime.Clock()

	response := models.RouteResponse{
		Routes: []models.Route{
			{
				Summary: models.RouteSummary{
					LengthInMeters:                   int(distance),
					TravelTimeInSeconds:              int(travelTimeInSeconds),
					TrafficDelayInSeconds:            0,
					TrafficLengthInMeters:            0,
					DepartureTime:                    time.Date(year, month, day, hour, minute, second, 0, time.Local),
					ArrivalTime:                      time.Date(endYear, endMonth, endDay, endHour, endMinute, endSecond, 0, time.Local),
					ClearanceRequired:                clearanceRequired,
					RemainingOperationTimeAtLocation: int(remainingOperationTimeAtLocation),
					ClearanceZones:                   clearenceZonesCrossed,
				},
				Legs: []models.Leg{
					{
						Summary: models.LegSummary{
							LengthInMeters:                   int(distance),
							TravelTimeInSeconds:              int(travelTimeInSeconds),
							TrafficDelayInSeconds:            0,
							TrafficLengthInMeters:            0,
							DepartureTime:                    time.Date(year, month, day, hour, minute, second, 0, time.Local),
							ArrivalTime:                      time.Date(endYear, endMonth, endDay, endHour, endMinute, endSecond, 0, time.Local),
							ClearanceRequired:                clearanceRequired,
							RemainingOperationTimeAtLocation: int(remainingOperationTimeAtLocation),
							ClearanceZones:                   clearenceZonesCrossed,
						},
						Points: waypoints,
					},
				},
			},
		},
	}

	// Return the JSON response
	return c.JSON(http.StatusOK, response)
}

func calculateTime(distance float64) float64 {
	speed := 1.0 // speed in 1meter/second
	time := distance / speed
	return time // time in seconds
}

func bindParamQuery(c echo.Context) (string, error) {
	query := c.QueryParam("query")
	if query != "" {
		return c.QueryParam("query"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

func bindParamRouteType(c echo.Context) (string, error) {
	routeType := c.QueryParam("routeType")
	if routeType != "" {
		return c.QueryParam("routeType"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

func bindParamTravelMode(c echo.Context) (string, error) {
	travelMode := c.QueryParam("travelMode")
	if travelMode != "" {
		return c.QueryParam("travelMode"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

func bindParamRouteRepresentation(c echo.Context) (string, error) {
	routeRepresentation := c.QueryParam("routeRepresentation")
	if routeRepresentation != "" {
		return c.QueryParam("routeRepresentation"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

func bindParamComputeBestOrder(c echo.Context) (string, error) {
	computeBestOrder := c.QueryParam("computeBestOrder")
	if computeBestOrder != "" {
		return c.QueryParam("computeBestOrder"), nil
	}
	return "", fmt.Errorf("error in computeBestOrder")
}

func bindParamResourceId(c echo.Context) (string, error) {
	resourceId := c.QueryParam("resourceId")
	if resourceId != "" {
		return c.QueryParam("query"), nil
	}
	return "", fmt.Errorf("error in bindDroneIdParam")
}

/*
* Errors functions
TODO Find a way to optimize this function
*/
func handleErrors(c echo.Context, ID string, err error) error {
	if strings.Contains(err.Error(), "Unknown id") || strings.Contains(err.Error(), "Value too long for type") {
		return handleBadRequest(c, err)
	}
	if strings.Contains(err.Error(), "connect: network is unreachable") {
		return handleInternalError(c, fmt.Errorf("connection to Drone has failed: network is unreachable"))
	}
	return handleInternalError(c, fmt.Errorf("Error during request: "+ID))
}

func handleInternalError(c echo.Context, err error) error {
	return c.JSON(http.StatusInternalServerError, err.Error())
}

func handleBadRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, err.Error())
}

func (co *Emulator) Dispose() error {
	// Nothing to do
	service.Dispose()
	return nil
}
