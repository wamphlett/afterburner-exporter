package config

import "gopkg.in/ini.v1"

type MQTTConfig struct {
	Broker string
	Port   int
}

func NewMQTTConfigFromFile(section *ini.Section) *MQTTConfig {
	port, err := section.Key("port").Int()
	if err != nil {
		panic("invalid MQTT port")
	}
	cfg := &MQTTConfig{
		Broker: section.Key("broker").String(),
		Port:   port,
	}
	if cfg.isValid() {
		return cfg
	}
	return nil
}

func (i *MQTTConfig) isValid() bool {
	if i.Broker == "" || i.Port == 0 {
		return false
	}
	return true
}
