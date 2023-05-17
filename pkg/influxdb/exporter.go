package influxdb

import (
	"context"
	"log"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/wamphlett/afterburner-exporter/config"
)

type Exporter struct {
	sync.Mutex
	influx       influxdb2.Client
	writer       api.WriteAPIBlocking
	currentBatch []*write.Point
}

func New(cfg *config.InfluxDB2Config) *Exporter {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &Exporter{
		influx:       client,
		writer:       client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
		currentBatch: []*write.Point{},
	}
}

func (i *Exporter) AddToBatch(device, field string, value float64, timestamp time.Time) error {
	i.Lock()
	defer i.Unlock()
	p := influxdb2.NewPoint("afterburner", map[string]string{"device": device}, map[string]interface{}{field: value}, timestamp)
	i.currentBatch = append(i.currentBatch, p)
	return nil
}

func (i *Exporter) Flush() error {
	i.Lock()
	defer i.Unlock()
	if err := i.writer.WritePoint(context.TODO(), i.currentBatch...); err != nil {
		log.Printf("failed to write points to InfluxDB2: %s", err.Error())
	}
	// empty the current batch
	i.currentBatch = []*write.Point{}
	return nil
}
