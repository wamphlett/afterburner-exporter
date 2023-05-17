package config

import "gopkg.in/ini.v1"

type MQTTConfig struct {
	Topic  string
	Broker string
	Port   int
}

func NewMQTTConfigFromFile(section *ini.Section) *MQTTConfig {
	port, err := section.Key("port").Int()
	if err != nil {
		panic("invalid MQTT port")
	}
	cfg := &MQTTConfig{
		Topic: section.Key("topic").String(),,
		Broker: section.Key("broker").String(),
		Port:   port,
	}
	if cfg.isValid() {
		return cfg
	}
	return nil
}

func (i *MQTTConfig) isValid() bool {
	if i.Broker == "" || i.Topic == "" || i.Port == 0 {
		return false
	}
	return true
}
