package main

import (
	"fmt"
	"time"
)

type Exporter interface {
	AddToBatch(device, field string, value interface{}, timestamp time.Time) error
	Flush() error
}

func process(file string, exporters []Exporter) {
	fmt.Println("Processing file", file)
}
