package mqtt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wamphlett/afterburner-exporter/config"
)

type aggregator struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Last  float64 `json:"last"`
	Count int     `json:"count"`
	Mean  float64 `json:"mean"`
	sum   float64
}

func (a *aggregator) register(v float64) {
	if v < a.Min || a.Min == 0 {
		a.Min = v
	}

	if v > a.Max {
		a.Max = v
	}

	a.Last = v
	a.Count++
	a.sum += v
	a.Mean = a.sum / float64(a.Count)
}

func (a *aggregator) calculateAverage() {
	a.Mean = a.sum / float64(a.Count)
}

type Exporter struct {
	sync.Mutex
	client  mqtt.Client
	topic   string
	metrics map[string]*aggregator
}

func New(cfg *config.MQTTConfig) *Exporter {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Broker, cfg.Port))
	opts.SetClientID("afterburner_exporter")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return &Exporter{
		metrics: make(map[string]*aggregator),
		client:  client,
		topic:   cfg.Topic,
	}
}

func (e *Exporter) AddToBatch(_, field string, value float64, _ time.Time) error {
	if _, ok := e.metrics[field]; !ok {
		e.metrics[field] = &aggregator{}
	}
	e.metrics[field].register(value)
	return nil
}

func (e *Exporter) Flush() error {
	e.Lock()
	defer e.Unlock()
	defer func() {
		e.metrics = make(map[string]*aggregator)
	}()

	bytes, err := json.Marshal(e.metrics)
	if err != nil {
		return err
	}

	_ = e.client.Publish(e.topic, 0, false, bytes)
	return nil
}
