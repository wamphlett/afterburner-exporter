package processor

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Exporter interface {
	AddToBatch(device, field string, value interface{}, timestamp time.Time) error
	Flush() error
}

type Processor struct {
	file      string
	exporters []Exporter
	stop      chan bool
}

func New(file string, interval time.Duration, exporters []Exporter) *Processor {
	p := &Processor{
		file:      file,
		exporters: exporters,
		stop:      make(chan bool),
	}

	go p.start(interval)

	return p
}

func (p *Processor) start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			p.process()
		}
	}
}

func (p *Processor) process() {
	// if there is no log file, return without doing anything
	if _, err := os.Stat(p.file); errors.Is(err, os.ErrNotExist) {
		return
	}

	log.Printf("processing file: %s", p.file)

	// rename the file to stop afterburner writing to it
	lockedFilePath := p.file + ".locked"
	e := os.Rename(p.file, lockedFilePath)
	if e != nil {
		log.Fatal(e)
	}

	defer func() {
		if err := os.Remove(lockedFilePath); err != nil {
			log.Fatal(err)
		}
	}()

	f, err := os.Open(lockedFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	device := ""
	fields := []string{}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		switch record[0] {
		// capture the device name
		case "01":
			device = strings.TrimSpace(record[2])
		// reset headers
		case "02":
			fields = []string{}
			for i, value := range record {
				if i < 2 {
					continue
				}
				fields = append(fields, strings.ToLower(strings.TrimSpace(value)))
			}
		// extract the values
		case "80":
			layout := "02-01-2006 15:04:05"
			location, err := time.LoadLocation("Europe/London")
			if err != nil {
				location = time.UTC
			}
			timestamp, _ := time.ParseInLocation(layout, strings.TrimSpace(record[1]), location)
			for i, field := range fields {
				// ignore any value which cant be parsed as a float
				value := strings.TrimSpace(record[i+2])
				parsedValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					if value != "N/A" {
						log.Printf("skipping value which cannot be parsed: %s", value)
					}
					continue
				}
				// add the data to each exporter
				for _, exporter := range p.exporters {
					_ = exporter.AddToBatch(device, field, parsedValue, timestamp)
				}
			}
		}
	}

	// flush all the exporters
	for _, exporter := range p.exporters {
		if err = exporter.Flush(); err != nil {
			log.Printf("failed to flush batch: %s", err.Error())
		}
	}
}

func (p *Processor) Stop() {
	p.stop <- true
}
