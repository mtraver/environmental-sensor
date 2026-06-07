// Program iotcorelogger reads from sensors and publishes the measurements to AWS IoT Core over MQTT.
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/host/v3"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	flagAWSDeviceFilePath string
	flagPort              int
	flagDryrun            bool
)

func init() {
	flag.StringVar(&flagAWSDeviceFilePath, "aws-device", "", "path to a device config file describing an AWS IoT Core device")
	flag.IntVar(&flagPort, "port", 8080, "port on which the device's web server should listen")
	flag.BoolVar(&flagDryrun, "dryrun", false, "set to true to print rather than publish measurements")

	flag.Usage = func() {
		message := `usage: iotcorelogger [options]

Options:
`

		fmt.Fprint(flag.CommandLine.Output(), message)
		flag.PrintDefaults()
	}
}

func parseFlags() error {
	flag.Parse()

	if flagAWSDeviceFilePath == "" {
		return errors.New("-aws-device must be given")
	}

	return nil
}

func startHTTPServer(mux *http.ServeMux) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", flagPort),
		Handler: mux,
	}

	go func() {
		log.Printf("Web server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	return srv
}

func main() {
	if err := parseFlags(); err != nil {
		fmt.Printf("argument error: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	templates := template.Must(template.New("index").ParseFS(templatesFS, "templates/*.html"))

	// Parse device file.
	device, err := parseDeviceFile(flagAWSDeviceFilePath)
	if err != nil {
		log.Fatalf("Failed to parse AWS device file: %v", err)
	}

	// Initialize periph.
	if _, err := host.Init(); err != nil {
		log.Fatalf("Failed to initialize periph: %v", err)
	}

	// We'll run until cancelled by the user (e.g. ctrl-c).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	monitor, err := NewMonitor(ctx, device)
	if err != nil {
		log.Fatal(err)
	}

	// Start up a web server that provides basic info about the device.
	mux := http.NewServeMux()
	mux.Handle("/{$}", &rootHandler{
		templates: templates,
		mon:       monitor,
	})
	srv := startHTTPServer(mux)

	<-ctx.Done()

	// Shut down the HTTP server.
	log.Println("Shutting down HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during HTTP server shutdown: %v", err)
	}

	// Clean up the Monitor's resources, including disconnecting from the MQTT broker.
	log.Println("Cleaning up resources and disconnecting from MQTT broker...")
	closeCtx, closeCancel := context.WithTimeout(context.Background(), timeout)
	defer closeCancel()
	if err := monitor.Close(closeCtx); err != nil {
		log.Printf("Error during cleanup or disconnect: %v", err)
	}

	log.Println("Done")
}
