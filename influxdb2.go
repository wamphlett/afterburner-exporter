package main

import (
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxDB2Client struct {
	influx       influxdb2.Client
	writer       api.WriteAPIBlocking
	currentBatch []*write.Point
	mutex        sync.Mutex
}

func NewInfluxDB2Client(cfg *InfluxDB2Config) *InfluxDB2Client {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &InfluxDB2Client{
		influx:       client,
		writer:       client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
		currentBatch: []*write.Point{},
	}
}

func (i *InfluxDB2Client) AddToBatch(device, field string, value interface{}, timestamp time.Time) error {
	i.mutex.Lock()
	defer i.mutex.Lock()
	p := influxdb2.NewPoint("afterburner", map[string]string{}, map[string]interface{}{field: value}, timestamp)
	i.currentBatch = append(i.currentBatch, p)
	return nil
}

func (i *InfluxDB2Client) Flush() error {
	i.mutex.Lock()
	defer i.mutex.Lock()
	return nil
}
