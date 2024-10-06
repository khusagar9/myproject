package main

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

	"h3d-drone-emulator/config"
	"h3d-drone-emulator/controller"
	"h3d-drone-emulator/service"

	commonController "gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/controller"
	"gitlab.thalesdigital.io/prs-sdp/shared/libs/golang/sdp-common-backend.git/log"
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

	if GitCommit != "" {
		log.Info("GitCommit : [%s]", GitCommit)
	}

	backendID := "h3d-drone-emulator"
	log.Info("Starting [%s]", backendID)
	log.Info(".................................. H3D DroneSubSystem Successfully Started ..................................")

	log.Info("Setting up Echo")
	e := commonController.CreateEcho()

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
