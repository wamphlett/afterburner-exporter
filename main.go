package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serviceName = "Afterburner Exporter"
	serviceDesc = "Exports Afterburner monitoring files to InfluxDB"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg := NewConfigFromFile()
	fmt.Println(cfg.InfluxDB2)
	exporters := []Exporter{}
	if cfg.InfluxDB2 != nil {
		exporters = append(exporters, NewInfluxDB2Client(cfg.InfluxDB2))
	}

	ticker := time.NewTicker(cfg.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				process(cfg.File, exporters)
			}
		}
	}()

	// wait for signal
	<-sigs
}
