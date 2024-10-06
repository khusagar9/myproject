package config

import (
	"flag"

	"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/log"
	sdpS3 "gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/s3"
)

type AppConfig struct {
	ServerPort          *int
	EndPointUrl         *string
	VersionPath         *string
	Topic               *string
	BrokerAddress       *string
	RestAPIAddress      *string
	KafkaLocationTopic  *string
	KafkaGwAddress      *string
	KeycloakTokenUrl    *string
	GetHealthPath       *string
	PublishLocPath      *string
	DroneInfoPath       *string
	DroneVideoPath      *string
	ResourcesBasePath   *string
	RmsRestAPIAddress   *string
	RmsResourceBasePath *string
	RmsSquadsBasePath   *string
	DroneBasePath       *string
	AllDronesPath       *string
	AllFlightsPath      *string
	// StartMissionPath    *string
	// StopMissionPath     *string
	GetMissionPath      *string
	DroneServerPath     *string
	AllDroneServersPath *string
	VideoFeedPath       *string
	DroneSpeed          *int
	Altitude            *int
	Temperature         *int
	SignalStrength      *string
	BatteryLife         *float64
	StartingLat         *float64
	StartingLong        *float64
	DroneIds            *string
	SimulateRealH3D     *bool
	LogLevel            *string
	S3Config            sdpS3.S3Config
	NoFlyZoneEndPoint   *string
	GetRoutePath        *string
	DispatchTime        *int
	ClearanceTime       *int
}

var appConfig AppConfig

func Load() {
	appConfig = AppConfig{
		ServerPort:          flag.Int("server-port", 11000, "Rest Server port"),
		EndPointUrl:         flag.String("endpoint-url", "/h3d-drone-emulator", "Endpoint Url of Drone Emulator"),
		VersionPath:         flag.String("version-path", "/v0", "Version Path of Endpoint"),
		Topic:               flag.String("drone-stats-topic", "dronestats", "Drone stats topic"),
		BrokerAddress:       flag.String("broker-address", "localhost:9092", "Broker Address"),
		RestAPIAddress:      flag.String("drone-connector-url", "http://127.0.0.1:8077/drone-connector/v0", "Url of Drone Connector"),
		KafkaLocationTopic:  flag.String("kafka-location-topic", "RmsResourceLocation", "Kakfa Location Topic"),
		KafkaGwAddress:      flag.String("kafka-gw-address", "https://pilot.sdpcore.apps.thalesdigital.io/rms/v0", "Kafka Gateway Address"),
		KeycloakTokenUrl:    flag.String("keycloak-token-url", "http://keycloak-http.authentication/auth/realms/sdp/protocol/openid-connect/token", "Keycloak Token Url"),
		GetHealthPath:       flag.String("get-health-path", "/health", "Get Health Path"),
		PublishLocPath:      flag.String("publish-loc-path", "/postlocation", "Publish Location Path"),
		ResourcesBasePath:   flag.String("resources-base-path", "/resources", "Resources Base Path"),
		RmsRestAPIAddress:   flag.String("rms-url", "http://sdp-rms-go.sdp-gateways:8080/rms/v1/", "Url of RMS endpoints"),
		RmsResourceBasePath: flag.String("rms-resource-base-path", "/resource", "RMS Resource Base Path"),
		RmsSquadsBasePath:   flag.String("rms-squads-base-path", "/squads", "RMSResource Base Path"),
		DroneBasePath:       flag.String("drone-base-path", "/drone", "Drone Base Path"),
		DroneInfoPath:       flag.String("drone-info-path", "/status", "Drone Info Path"),
		DroneVideoPath:      flag.String("drone-video-path", "/video", "Drone Video Path"),
		AllDronesPath:       flag.String("all-drones-path", "/drones", "All Drones Path"),
		AllFlightsPath:      flag.String("all-flights-path", "/flights", "All Flights Path"),
		// StartMissionPath:    flag.String("start-mission-path", "/flynow", "Start Mission Path"),
		// StopMissionPath:     flag.String("stop-mission-path", "/cancel", "Stop Mission Path"),
		GetMissionPath:      flag.String("get-mission-path", "/mission/", "Get Mission Path"),
		DroneServerPath:     flag.String("drone-server-path", "/dbx/{dbx_id}/read", "Get Drones Server Path"),
		AllDroneServersPath: flag.String("all-drones-servers-path", "/dbxs", "All Drones Servers Path"),
		VideoFeedPath:       flag.String("video-feed-path", "/dbx/{dbx_id}/video", "Video Feed Path"),
		DroneSpeed:          flag.Int("drone-speed", 67, "Drone speed in miles per hour"),
		Altitude:            flag.Int("altitude", 400, "Drone altitude in feet"),
		Temperature:         flag.Int("temperature", 31, "Temperature in degrees Celsius"),
		SignalStrength:      flag.String("signal-strength", "Excellent", "Signal strength of drone"),
		BatteryLife:         flag.Float64("battery-life", 30, "Battery life in minutes"),
		StartingLat:         flag.Float64("starting-lat", 1.333558, "Starting laitude of Drone"),
		StartingLong:        flag.Float64("starting-long", 103.816614, "Starting longitude of Drone"),
		DroneIds:            flag.String("drone-ids", "605d5aa3c9f9e6b0e44a2925,60501d53f576cd66a4f2b,60501d53f576cd66a42c,6051b9c144811f30c5902ee6", "Drone Ids to Emulate separated by ,"),
		SimulateRealH3D:     flag.Bool("simulate-real-h3d", true, "Simulate Real H3D option"),
		LogLevel:            flag.String("log-level", "info", "The log level of the application."),
		S3Config: sdpS3.S3Config{
			Endpoint:        flag.String("s3-endpoint", "https://pilot.sandbox.sdpcore.apps.thalesdigital.io/", "Endpoint for S3 service"),
			Region:          flag.String("s3-region", "us-east-1", "S3 region"),
			AccessKeyID:     flag.String("s3-access-key", "sdp-backend", "S3 access key"),
			SecretAccessKey: flag.String("s3-secret-key", "password", "S3 secret key"),
			DisableSSL:      flag.Bool("s3-disable-ssl", true, "True if SSL should be disable when connecting to S3 storage"),
			ForcePathStyle:  flag.Bool("s3-force-path-style", true, "True to force path style URL when using S3 APIs"),
		},
		NoFlyZoneEndPoint: flag.String("no-fly-zones-url", "https://pilot.sdpcore.apps.thalesdigital.io/custom_entity/v0/internal/instances/search", "Url for No Fly Zones"),
		GetRoutePath:      flag.String("get-route-path", "/route", "Get Route Path"),
		DispatchTime:      flag.Int("dispatchTime", 60, "Dispatch time in seconds"),
		ClearanceTime:     flag.Int("clearanceTime", 300, "Clearance time in seconds"),
	}

	flag.Parse()
	log.SetLevel(*appConfig.LogLevel)

}

// Get Application configuration
func Get() AppConfig {
	return appConfig
}
