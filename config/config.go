package config

import (
	"time"

	"gopkg.in/ini.v1"
)

type Config struct {
	File      string
	Interval  time.Duration
	InfluxDB2 *InfluxDB2Config
}

type InfluxDB2Config struct {
	URL    string
	Org    string
	Bucket string
	Token  string
}

func (i *InfluxDB2Config) isValid() bool {
	for _, v := range []string{i.URL, i.Org, i.Bucket, i.Token} {
		if v == "" {
			return false
		}
	}
	return true
}

func withDefaultConfigValues() *Config {
	return &Config{
		Interval: time.Millisecond * 15000,
	}
}

func NewConfigFromFile() *Config {
	cfg := withDefaultConfigValues()

	cfgFile, err := ini.Load("exporter.conf")
	if err != nil {
		return cfg
	}

	if file := cfgFile.Section("").Key("log_path").String(); file != "" {
		cfg.File = file
	}

	if interval, err := cfgFile.Section("").Key("interval").Int64(); err == nil {
		cfg.Interval = time.Millisecond * time.Duration(interval)
	}

	// If there is an InfluxDB2 config, set it up
	if cfgFile.HasSection("influxdb2") {
		cfg.InfluxDB2 = NewInfluxDB2ConfigFromFile(cfgFile.Section("influxdb2"))

	}

	return cfg
}

func NewInfluxDB2ConfigFromFile(section *ini.Section) *InfluxDB2Config {
	cfg := &InfluxDB2Config{
		URL:    section.Key("url").String(),
		Org:    section.Key("org").String(),
		Bucket: section.Key("bucket").String(),
		Token:  section.Key("token").String(),
	}
	if cfg.isValid() {
		return cfg
	}
	return nil
}
