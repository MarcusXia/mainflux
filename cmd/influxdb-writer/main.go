// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	//influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName = "influxdb-writer"

	defNatsURL    = "nats://localhost:4222"
	defLogLevel   = "error"
	defPort       = "8180"
	defDB         = "mainflux"
	defDBHost     = "localhost"
	defDBPort     = "8086"
	defDBUser     = "mainflux"
	defDBPass     = "mainflux"
	defConfigPath = "/config.toml"

	defDBBucket = "mainflux-bucket"
	defDBOrg    = "mainflux"
	defDBToken  = "mainflux-token"
	defDBUrl    = "http://localhost:8086"

	envNatsURL    = "MF_NATS_URL"
	envLogLevel   = "MF_INFLUX_WRITER_LOG_LEVEL"
	envPort       = "MF_INFLUX_WRITER_PORT"
	envDB         = "MF_INFLUXDB_DB"
	envDBHost     = "MF_INFLUXDB_HOST"
	envDBPort     = "MF_INFLUXDB_PORT"
	envDBUser     = "MF_INFLUXDB_ADMIN_USER"
	envDBPass     = "MF_INFLUXDB_ADMIN_PASSWORD"
	envConfigPath = "MF_INFLUX_WRITER_CONFIG_PATH"
	envDBBucket   = "MF_INFLUXDB_BUCKET"
	envDBOrg      = "MF_INFLUXDB_ORG"
	envDBToken    = "MF_INFLUXDB_TOKEN"
	envDBUrl      = "http://localhost:8086"
)

type config struct {
	natsURL    string
	logLevel   string
	port       string
	dbName     string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPass     string
	configPath string
	dbBucket   string
	dbOrg      string
	dbToken    string
	dbUrl      string
}

func main() {
	cfg /*, clientCfg*/ := loadConfigs()

	println("Hello from influxdb Writer")
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	client, err := connectToInfluxdb(cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	println("Connected to INFLUXDB2!")
	defer client.Close()

	//counter, latency := makeMetrics()
	// repo = api.LoggingMiddleware(repo, logger)
	//repo = api.MetricsMiddleware(repo, counter, latency)

	//if err := consumers.Start(pubSub, repo, cfg.configPath, logger); err != nil {
	//	logger.Error(fmt.Sprintf("Failed to start InfluxDB writer: %s", err))
	//	os.Exit(1)
	//}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))

}

func connectToInfluxdb(cfg config) (influxdb2.Client, error) {
	// token = Q8uRqtnzr2O-RZlgavoB86GR1-yLBjA0K762HZU1jU9fG__Scu7A7eb8YOIjzdvplCWZRcs5wIVI5FgtAl-0fg==
	// I can see this token when I open the UI. but I cannot get health as Expected.

	client := influxdb2.NewClient(cfg.dbUrl, cfg.dbToken)
	println("client instance created")
	_, err := client.Health(context.Background())
	return client, err
}

func loadConfigs() config /*influxdata.HTTPConfig*/ {
	cfg := config{
		natsURL:    mainflux.Env(envNatsURL, defNatsURL),
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		port:       mainflux.Env(envPort, defPort),
		dbName:     mainflux.Env(envDB, defDB),
		dbHost:     mainflux.Env(envDBHost, defDBHost),
		dbPort:     mainflux.Env(envDBPort, defDBPort),
		dbUser:     mainflux.Env(envDBUser, defDBUser),
		dbPass:     mainflux.Env(envDBPass, defDBPass),
		configPath: mainflux.Env(envConfigPath, defConfigPath),
		dbBucket:   mainflux.Env(envDBBucket, defDBBucket),
		dbOrg:      mainflux.Env(envDBOrg, defDBOrg),
		dbToken:    mainflux.Env(envDBToken, defDBToken),
		dbUrl:      mainflux.Env(envDBUrl, defDBUrl),
	}
	/*
		clientCfg := influxdata.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort),
			Username: cfg.dbUser,
			Password: cfg.dbPass,
		}
	*/
	return cfg //, clientCfg
}

func makeMetrics() (*kitprometheus.Counter, *kitprometheus.Summary) {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_count",
		Help:      "Number of database inserts.",
	}, []string{"method"})

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of inserts in microseconds.",
	}, []string{"method"})

	return counter, latency
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("InfluxDB writer service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, api.MakeHandler(svcName))
}
