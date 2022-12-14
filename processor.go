package main

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

func process(file string, exporters []Exporter) {
	// if there is no log file, return without doing anything
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return
	}

	log.Printf("processing file: %s", file)

	// rename the file to stop afterburner writing to it
	lockedFilePath := file + ".locked"
	e := os.Rename(file, lockedFilePath)
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
				for _, exporter := range exporters {
					_ = exporter.AddToBatch(device, field, parsedValue, timestamp)
				}
			}
		}
	}

	// flush all the exporters
	for _, exporter := range exporters {
		if err = exporter.Flush(); err != nil {
			log.Printf("failed to flush batch: %s", err.Error())
		}
	}
}
