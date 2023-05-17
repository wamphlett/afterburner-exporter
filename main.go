package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wamphlett/afterburner-exporter/config"
	"github.com/wamphlett/afterburner-exporter/pkg/influxdb"
	"github.com/wamphlett/afterburner-exporter/pkg/processor"
)

const (
	serviceName = "Afterburner Exporter"
	serviceDesc = "Exports Afterburner monitoring files to InfluxDB"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg := config.NewConfigFromFile()
	fmt.Println(cfg.InfluxDB2)
	exporters := []processor.Exporter{}
	if cfg.InfluxDB2 != nil {
		exporters = append(exporters, influxdb.New(cfg.InfluxDB2))
	}

	// create a processor
	p := processor.New(cfg.File, cfg.Interval, exporters)

	// wait for signal
	<-sigs
	p.Stop()
	fmt.Println("stopping exporter")
}
