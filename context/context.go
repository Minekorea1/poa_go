package context

import (
	"poa/jsonWrapper"
)

type Context struct {
	Version string

	Configs Configs

	// UpdateAddress          string
	// UpdateCheckIntervalSec int
	// MqttBrokerAddress      string
	// MqttPort               int
	// PoaIntervalSec         int
}

type Configs struct {
	UpdateAddress          string
	UpdateCheckIntervalSec int
	MqttBrokerAddress      string
	MqttPort               int
	PoaIntervalSec         int
}

type DeviceType int

const (
	DeviceTypeNormal DeviceType = iota
	DeviceTypeDeeper
)

func NewContext() *Context {
	context := Context{
		// Configs: Configs{},
	}
	context.Configs.ReadFile("config.json")
	return &context
}

func (configs *Configs) ToJson() string {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	if jsonConfig.MarshalValue(configs) {
		return jsonConfig.ToString()
	}
	return ""
}

func (configs *Configs) ReadFile(path string) {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	jsonConfig.ReadJsonTo(path, configs)
}

func (configs *Configs) WriteFile(path string) {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	if jsonConfig.MarshalValue(configs) {
		jsonConfig.WriteJson(path)
	}
}
