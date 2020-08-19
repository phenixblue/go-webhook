package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"tkev.dev/imageswap-webhook/pkg/logger"
)

// EndpointLogger export a global Endpoint Logger instance
var EndpointLogger *zap.Logger = logger.InitLogger()

func rootHandler(w http.ResponseWriter, r *http.Request) {

	EndpointLogger.Info("Endpoint called", zap.String("path", r.URL.Path))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"alive": true}`)
	EndpointLogger.Info("Endpoint called", zap.String("path", r.URL.Path))

}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

	//statusCode := strconv.Itoa(r.Response.StatusCode)

	EndpointLogger.Info(
		"Endpoint called",
		zap.String("path", r.URL.Path),
		zap.String("test", strconv.Itoa(r.Response.StatusCode)),
	)
}

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	test := 5
	EndpointLogger.Info("Starting ImageSwap Webhook", zap.String("path", "NA"), zap.Int("code", test))

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler).Methods("GET", "POST")
	r.HandleFunc("/healthz", healthHandler).Methods("GET")
	r.HandleFunc("/metricsz", metricsHandler).Methods("GET")
	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:5000",
		// Set generic timeouts
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			zap.S().Error(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// Capture SIGINT anfd SIGTERM for graceful shutdowns
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	EndpointLogger.Info("Shutting down")
	os.Exit(0)

}
